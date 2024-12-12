package mail

import (
	"bytes"
	"context"
	"html/template"

	"fmt"
	"net/smtp"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/templates"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/k3a/html2text"
	"github.com/mailgun/mailgun-go/v4"
)

type MailTemplate struct {
	Mail   types.Email
	Domain string
}

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
func SendHTMLMail(to, subject string, msg types.Email, attachment []types.EmailAttachment) error {
	var renderer = templates.GetMailTemplate()

	var err error
	var body bytes.Buffer

	if utils.Config.Frontend.Mail.SMTP.User != "" {
		headers := "MIME-version: 1.0;\nContent-Type: text/html;"
		body.Write([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n%s\r\n", to, subject, headers)))
		err = renderer.Execute(&body, MailTemplate{Mail: msg, Domain: utils.Config.Frontend.SiteDomain})
		if err != nil {
			return fmt.Errorf("error rendering mail template: %w", err)
		}

		log.Infof("Email Attachments will not work with SMTP server")
		err = SendMailSMTP(to, body.Bytes())
	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
		err = renderer.ExecuteTemplate(&body, "layout", MailTemplate{Mail: msg, Domain: utils.Config.Frontend.SiteDomain})
		if err != nil {
			log.Error(err, "error rendering mail template", 0)
		}
		content := body.String()
		err = SendMailMailgun(to, subject, content, createTextMessage(msg), attachment)
	} else {
		log.Error(nil, "error sending reset-email: invalid config for mail-service", 0)
		err = nil
	}
	return err
}

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
func SendTextMail(to, subject, msg string, attachment []types.EmailAttachment) error {
	var err error
	if utils.Config.Frontend.Mail.SMTP.User != "" {
		log.Infof("Email Attachments will not work with SMTP server")
		err = SendTextMailSMTP(to, subject, msg)
	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
		err = SendTextMailMailgun(to, subject, msg, attachment)
	} else {
		err = fmt.Errorf("invalid config for mail-service")
	}
	return err
}

func createTextMessage(msg types.Email) string {
	return fmt.Sprintf("%s\n\n%s\n\n― You are receiving this because you are registered on beaconcha.in. You can manage your subscriptions at %s.", msg.Title, msg.Body, msg.SubscriptionManageURL)
}

// SendMailRateLimited sends an email to a given address with the given message.
// It will return a ratelimit-error if the configured ratelimit is exceeded.
func SendMailRateLimited(content types.TransitEmailContent, maxEmailsPerDay int64, bucket string) error {
	sendThresholdReachedMail := false
	count, err := db.CountSentMessage(bucket, content.UserId)
	if err != nil {
		return err
	}
	timeLeft := time.Until(time.Now().Add(utils.Day).Truncate(utils.Day))

	log.Debugf("user %d has sent %d of %d emails today, time left is %v", content.UserId, count, maxEmailsPerDay, timeLeft)
	if count > maxEmailsPerDay {
		return &types.RateLimitError{TimeLeft: timeLeft}
	} else if count == maxEmailsPerDay {
		sendThresholdReachedMail = true
	}

	err = SendHTMLMail(content.Address, content.Subject, content.Email, content.Attachments)
	if err != nil {
		log.Error(err, "error sending email", 0)
	}

	// make sure the threshold reached email arrives last
	if sendThresholdReachedMail {
		// send an email if this was the last email for today
		err := SendHTMLMail(content.Address,
			"beaconcha.in - Email notification threshold limit reached",
			types.Email{
				Title: "Email notification threshold limit reached",
				//nolint: gosec
				Body: template.HTML(fmt.Sprintf("You have reached the email notification threshold limit of %d emails per day. Further notification emails will be suppressed for %.1f hours.", maxEmailsPerDay, timeLeft.Hours())),
			},
			[]types.EmailAttachment{})
		if err != nil {
			return err
		}
	}

	return nil
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendMailSMTP(to string, msg []byte) error {
	server := utils.Config.Frontend.Mail.SMTP.Server // eg. smtp.gmail.com:587
	host := utils.Config.Frontend.Mail.SMTP.Host     // eg. smtp.gmail.com
	from := utils.Config.Frontend.Mail.SMTP.User     // eg. userxyz123@gmail.com
	password := utils.Config.Frontend.Mail.SMTP.Password
	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(server, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("error sending mail via smtp: %w", err)
	}

	return nil
}

// SendMailMailgun sends an email to the given address with the given message, using mailgun.
func SendMailMailgun(to, subject, msgHtml, msgText string, attachment []types.EmailAttachment) error {
	mg := mailgun.NewMailgun(
		utils.Config.Frontend.Mail.Mailgun.Domain,
		utils.Config.Frontend.Mail.Mailgun.PrivateKey,
	)

	// if the text part still contains html tags / entities, remove / convert them
	msgText = html2text.HTML2Text(msgText)

	message := mg.NewMessage(utils.Config.Frontend.Mail.Mailgun.Sender, subject, msgText, to)
	message.SetHtml(msgHtml)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if len(attachment) > 0 {
		for _, att := range attachment {
			message.AddBufferAttachment(att.Name, att.Attachment)
		}
	}

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		log.InfoWithFields(log.Fields{"resp": resp, "id": id}, "error sending mail via mailgun")
		return fmt.Errorf("error sending mail via mailgun: %w", err)
	}

	return nil
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendTextMailSMTP(to, subject, body string) error {
	server := utils.Config.Frontend.Mail.SMTP.Server // eg. smtp.gmail.com:587
	host := utils.Config.Frontend.Mail.SMTP.Host     // eg. smtp.gmail.com
	from := utils.Config.Frontend.Mail.SMTP.User     // eg. userxyz123@gmail.com
	password := utils.Config.Frontend.Mail.SMTP.Password
	auth := smtp.PlainAuth("", from, password, host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	err := smtp.SendMail(server, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("error sending mail via smtp: %w", err)
	}

	return nil
}

// SendMailMailgun sends an email to the given address with the given message, using mailgun.
func SendTextMailMailgun(to, subject, msg string, attachment []types.EmailAttachment) error {
	mg := mailgun.NewMailgun(
		utils.Config.Frontend.Mail.Mailgun.Domain,
		utils.Config.Frontend.Mail.Mailgun.PrivateKey,
	)
	message := mg.NewMessage(utils.Config.Frontend.Mail.Mailgun.Sender, subject, msg, to)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if len(attachment) > 0 {
		for _, att := range attachment {
			message.AddBufferAttachment(att.Name, att.Attachment)
		}
	}

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		log.InfoWithFields(log.Fields{"resp": resp, "id": id}, "error sending mail via mailgun")
		return fmt.Errorf("error sending mail via mailgun: %w", err)
	}

	return nil
}
