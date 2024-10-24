package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/mail"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func InitNotificationSender() {
	log.Infof("starting notifications-sender")
	go notificationSender()
}

func notificationSender() {
	for {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)

		conn, err := db.FrontendWriterDB.Conn(ctx)
		if err != nil {
			log.Error(err, "error creating connection", 0)
			cancel()
			continue
		}

		_, err = conn.ExecContext(ctx, `SELECT pg_advisory_lock(500)`)
		if err != nil {
			log.Error(err, "error getting advisory lock from db", 0)

			err := conn.Close()
			if err != nil {
				log.Error(err, "error returning connection to connection pool", 0)
			}
			cancel()
			continue
		}

		log.Infof("lock obtained")
		err = dispatchNotifications()
		if err != nil {
			log.Error(err, "error dispatching notifications", 0)
		}

		err = garbageCollectNotificationQueue()
		if err != nil {
			log.Error(err, "error garbage collecting notification queue", 0)
		}

		log.InfoWithFields(log.Fields{"duration": time.Since(start)}, "notifications dispatched and garbage collected")
		metrics.TaskDuration.WithLabelValues("service_notifications_sender").Observe(time.Since(start).Seconds())

		unlocked := false
		rows, err := conn.QueryContext(ctx, `SELECT pg_advisory_unlock(500)`)
		if err != nil {
			log.Error(err, "error executing advisory unlock", 0)

			err = conn.Close()
			if err != nil {
				log.WarnWithStackTrace(err, "error returning connection to connection pool", 0)
			}
			cancel()
			continue
		}

		for rows.Next() {
			err = rows.Scan(&unlocked)
			if err != nil {
				log.Error(err, "error scanning advisory unlock result", 0)
			}
		}

		if !unlocked {
			log.Error(nil, fmt.Errorf("error releasing advisory lock unlocked: %v", unlocked), 0)
		}

		conn.Close()
		if err != nil {
			log.WarnWithStackTrace(err, "error returning connection to connection pool", 0)
		}
		cancel()

		services.ReportStatus("notification-sender", "Running", nil)
		time.Sleep(time.Second * 30)
	}
}

// garbageCollectNotificationQueue deletes entries from the notification queue that have been processed
func garbageCollectNotificationQueue() error {
	rows, err := db.WriterDb.Exec(`DELETE FROM notification_queue WHERE (sent < now() - INTERVAL '30 minutes') OR (created < now() - INTERVAL '1 hour')`)
	if err != nil {
		return fmt.Errorf("error deleting from notification_queue %w", err)
	}

	rowsAffected, _ := rows.RowsAffected()

	log.Infof("deleted %v rows from the notification_queue", rowsAffected)

	return nil
}

func dispatchNotifications() error {
	err := sendEmailNotifications()
	if err != nil {
		return fmt.Errorf("error sending email notifications, err: %w", err)
	}

	err = sendPushNotifications()
	if err != nil {
		return fmt.Errorf("error sending push notifications, err: %w", err)
	}

	err = sendWebhookNotifications()
	if err != nil {
		return fmt.Errorf("error sending webhook notifications, err: %w", err)
	}

	err = sendDiscordNotifications()
	if err != nil {
		return fmt.Errorf("error sending webhook discord notifications, err: %w", err)
	}

	return nil
}

func sendEmailNotifications() error {
	var notificationQueueItem []types.TransitEmail

	err := db.WriterDb.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue WHERE sent IS null AND channel = 'email' ORDER BY created ASC`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}

	log.Infof("processing %v email notifications", len(notificationQueueItem))

	for _, n := range notificationQueueItem {
		userInfo, err := db.GetUserInfo(context.Background(), uint64(n.Content.UserId), db.FrontendReaderDB)
		if err != nil {
			return err
		}
		err = mail.SendMailRateLimited(n.Content, int64(userInfo.PremiumPerks.EmailNotificationsPerDay), "n_emails")
		if err != nil {
			if !strings.Contains(err.Error(), "rate limit has been exceeded") {
				metrics.Errors.WithLabelValues("notifications_send_email").Inc()
				log.Error(err, "error sending email notification", 0)
			} else {
				metrics.NotificationsSent.WithLabelValues("email", "429").Inc()
			}
		} else {
			metrics.NotificationsSent.WithLabelValues("email", "200").Inc()
		}
		_, err = db.WriterDb.Exec(`UPDATE notification_queue set sent = now() where id = $1`, n.Id)
		if err != nil {
			return fmt.Errorf("error updating sent status for email notification with id: %v, err: %w", n.Id, err)
		}
	}
	return nil
}

func sendPushNotifications() error {
	var notificationQueueItem []types.TransitPush

	err := db.WriterDb.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue WHERE sent IS null AND channel = 'push' ORDER BY created ASC`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}

	log.Infof("processing %v push notifications", len(notificationQueueItem))

	batchSize := 500
	for _, n := range notificationQueueItem {
		for b := 0; b < len(n.Content.Messages); b += batchSize {
			start := b
			end := b + batchSize
			if len(n.Content.Messages) < end {
				end = len(n.Content.Messages)
			}

			err = SendPushBatch(n.Content.UserId, n.Content.Messages[start:end], false)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_send_push_batch").Inc()
				log.Error(err, "error sending firebase batch job", 0)
			} else {
				metrics.NotificationsSent.WithLabelValues("push", "200").Add(float64(len(n.Content.Messages)))
			}

			_, err = db.WriterDb.Exec(`UPDATE notification_queue SET sent = now() WHERE id = $1`, n.Id)
			if err != nil {
				return fmt.Errorf("error updating sent status for push notification with id: %v, err: %w", n.Id, err)
			}
		}
	}
	return nil
}

func sendWebhookNotifications() error {
	var notificationQueueItem []types.TransitWebhook

	err := db.WriterDb.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue WHERE sent IS null AND channel = 'webhook' ORDER BY created ASC`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}
	client := &http.Client{Timeout: time.Second * 30}

	log.Infof("processing %v webhook notifications", len(notificationQueueItem))

	for _, n := range notificationQueueItem {
		_, err := db.CountSentMessage("n_webhooks", n.Content.UserId)
		if err != nil {
			log.Error(err, "error counting sent webhook", 0)
		}

		// do not retry after 5 attempts
		if n.Content.Webhook.Retries > 5 {
			_, err := db.WriterDb.Exec(`DELETE FROM notification_queue WHERE id = $1`, n.Id)
			if err != nil {
				return fmt.Errorf("error deleting from notification queue: %w", err)
			}
			continue
		}

		reqBody := new(bytes.Buffer)

		err = json.NewEncoder(reqBody).Encode(n.Content)
		if err != nil {
			log.Error(err, "error marshalling webhook event", 0)
		}

		_, err = url.Parse(n.Content.Webhook.Url)
		if err != nil {
			_, err := db.WriterDb.Exec(`DELETE FROM notification_queue WHERE id = $1`, n.Id)
			if err != nil {
				return fmt.Errorf("error deleting from notification queue: %w", err)
			}
			continue
		}

		go func(n types.TransitWebhook) {
			if n.Content.Webhook.Retries > 0 {
				time.Sleep(time.Duration(n.Content.Webhook.Retries) * time.Second)
			}
			resp, err := client.Post(n.Content.Webhook.Url, "application/json", reqBody)
			if err != nil {
				log.Warnf("error sending webhook request: %v", err)
				metrics.NotificationsSent.WithLabelValues("webhook", "error").Inc()
				return
			} else {
				metrics.NotificationsSent.WithLabelValues("webhook", resp.Status).Inc()
			}
			defer resp.Body.Close()

			_, err = db.WriterDb.Exec(`UPDATE notification_queue SET sent = now() WHERE id = $1`, n.Id)
			if err != nil {
				log.Error(err, "error updating notification_queue table", 0)
				return
			}

			if resp != nil && resp.StatusCode < 400 {
				_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = 0, last_sent = now() WHERE id = $1;`, n.Content.Webhook.ID)
				if err != nil {
					log.Error(err, "error updating users_webhooks table", 0)
					return
				}
			} else {
				var errResp types.ErrorResponse

				if resp != nil {
					b, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error(err, "error reading body", 0)
					}

					errResp.Status = resp.Status
					errResp.Body = string(b)
				}

				_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = retries + 1, last_sent = now(), request = $2, response = $3 WHERE id = $1;`, n.Content.Webhook.ID, n.Content, errResp)
				if err != nil {
					log.Error(err, "error updating users_webhooks table", 0)
					return
				}
			}
		}(n)
	}
	return nil
}

func sendDiscordNotifications() error {
	var notificationQueueItem []types.TransitDiscord

	err := db.WriterDb.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue WHERE sent IS null AND channel = 'webhook_discord' ORDER BY created ASC`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}
	client := &http.Client{Timeout: time.Second * 30}

	log.Infof("processing %v discord webhook notifications", len(notificationQueueItem))
	webhookMap := make(map[uint64]types.UserWebhook)

	notifMap := make(map[uint64][]types.TransitDiscord)
	// generate webhook id => discord req
	// while mapping. aggregate embeds while doing so, up to 10 per req can be sent
	for _, n := range notificationQueueItem {
		// purge the event from existence if the retry counter is over 5
		if n.Content.Webhook.Retries > 5 {
			_, err = db.WriterDb.Exec(`DELETE FROM notification_queue where id = $1`, n.Id)
			if err != nil {
				log.Warnf("failed to delete notification from queue: %v", err)
			}
			continue
		}
		if _, exists := webhookMap[n.Content.Webhook.ID]; !exists {
			webhookMap[n.Content.Webhook.ID] = n.Content.Webhook
		}
		if _, exists := notifMap[n.Content.Webhook.ID]; !exists {
			notifMap[n.Content.Webhook.ID] = make([]types.TransitDiscord, 0)
		}
		notifMap[n.Content.Webhook.ID] = append(notifMap[n.Content.Webhook.ID], n)
	}
	for _, webhook := range webhookMap {
		// todo: this has the potential to spin up thousands of go routines
		// should use an errgroup instead if we decide to keep the aproach
		go func(webhook types.UserWebhook, reqs []types.TransitDiscord) {
			defer func() {
				// update retries counters in db based on end result
				if webhook.DashboardId == 0 && webhook.DashboardGroupId == 0 {
					_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = $1, last_sent = now() WHERE id = $2;`, webhook.Retries, webhook.ID)
				} else {
					_, err = db.WriterDb.Exec(`UPDATE users_val_dashboards_groups SET webhook_retries = $1, webhook_last_sent = now() WHERE id = $2 AND dashboard_id = $3;`, webhook.Retries, webhook.DashboardGroupId, webhook.DashboardId)
				}
				if err != nil {
					log.Warnf("failed to update retries counter to %v for webhook %v: %v", webhook.Retries, webhook.ID, err)
				}

				// mark notifcations as sent in db
				ids := make([]uint64, 0)
				for _, req := range reqs {
					ids = append(ids, req.Id)
				}
				_, err = db.WriterDb.Exec(`UPDATE notification_queue SET sent = now() where id = ANY($1)`, pq.Array(ids))
				if err != nil {
					log.Warnf("failed to update sent for notifcations in queue: %v", err)
				}
			}()

			_, err = url.Parse(webhook.Url)
			if err != nil {
				log.Error(err, "error parsing url", 0, log.Fields{"webhook_id": webhook.ID})
				return
			}

			for i := 0; i < len(reqs); i++ {
				if webhook.Retries > 5 {
					break // stop
				}
				// sleep between retries
				time.Sleep(time.Duration(webhook.Retries) * time.Second)

				reqBody := new(bytes.Buffer)
				err := json.NewEncoder(reqBody).Encode(reqs[i].Content.DiscordRequest)
				if err != nil {
					log.Error(err, "error marshalling discord webhook event", 0)
					continue // skip
				}

				log.Infof("sending discord webhook request to %s with: %v", webhook.Url, reqs[i].Content.DiscordRequest)
				resp, err := client.Post(webhook.Url, "application/json", reqBody)
				if err != nil {
					log.Warnf("failed sending discord webhook request %v: %v", webhook.ID, err)
					metrics.NotificationsSent.WithLabelValues("webhook_discord", "error").Inc()
				} else {
					metrics.NotificationsSent.WithLabelValues("webhook_discord", resp.Status).Inc()
				}
				if resp != nil && resp.StatusCode < 400 {
					webhook.Retries = 0
				} else {
					webhook.Retries++
					var errResp types.ErrorResponse

					if resp != nil {
						b, err := io.ReadAll(resp.Body)
						if err != nil {
							log.Error(err, "error reading body", 0)
						} else {
							errResp.Body = string(b)
						}
						errResp.Status = resp.Status
						resp.Body.Close()

						if resp.StatusCode != http.StatusOK {
							log.WarnWithFields(map[string]interface{}{"errResp.Body": utils.FirstN(errResp.Body, 1000), "webhook.Url": webhook.Url}, "error pushing discord webhook")
						}
						if webhook.DashboardId == 0 && webhook.DashboardGroupId == 0 {
							_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET request = $2, response = $3 WHERE id = $1;`, webhook.ID, reqs[i].Content.DiscordRequest, errResp)
						}
						if err != nil {
							log.Error(err, "error storing failure data in users_webhooks table", 0)
						}
					}

					i-- // retry, IMPORTANT to be at the END of the ELSE, otherwise the wrong index will be used in the commands above!
				}
			}
		}(webhook, notifMap[webhook.ID])
	}

	return nil
}

func SendTestEmail(ctx context.Context, userId types.UserId, dbConn *sqlx.DB) error {
	var email string
	err := dbConn.GetContext(ctx, &email, `SELECT email FROM users WHERE id = $1`, userId)
	if err != nil {
		return err
	}
	content := types.TransitEmailContent{
		UserId:  userId,
		Address: email,
		Subject: "Test Email",
		Email: types.Email{
			Title: "beaconcha.in - Test Email",
			Body:  "This is a test email from beaconcha.in",
		},
		Attachments: []types.EmailAttachment{},
		CreatedTs:   time.Now(),
	}
	err = mail.SendMailRateLimited(content, 10, "n_test_emails")
	if err != nil {
		return fmt.Errorf("error sending test email, err: %w", err)
	}

	return nil
}

func SendTestWebhookNotification(ctx context.Context, userId types.UserId, webhookUrl string, isDiscordWebhook bool) error {
	count, err := db.CountSentMessage("n_test_push", userId)
	if err != nil {
		return err
	}
	if count > 10 {
		return fmt.Errorf("rate limit has been exceeded")
	}

	client := http.Client{Timeout: time.Second * 5}

	if isDiscordWebhook {
		req := types.DiscordReq{
			Content: "This is a test notification from beaconcha.in",
		}
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(req)
		if err != nil {
			return fmt.Errorf("error marshalling discord webhook event: %w", err)
		}
		resp, err := client.Post(webhookUrl, "application/json", reqBody)
		if err != nil {
			return fmt.Errorf("error sending discord webhook request: %w", err)
		}
		defer resp.Body.Close()
	} else {
		// send a test webhook notification with the text "TEST" in the post body
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(`{data: "TEST"}`)
		if err != nil {
			return fmt.Errorf("error marshalling webhook event: %w", err)
		}
		resp, err := client.Post(webhookUrl, "application/json", reqBody)
		if err != nil {
			return fmt.Errorf("error sending webhook request: %w", err)
		}
		defer resp.Body.Close()
	}
	return nil
}
