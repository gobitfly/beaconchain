package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

const NOTIFICAION_EMAIL_RATE_LIMIT_BUCKET = "n_mails"
const NOTIFICAION_PUSH_RATE_LIMIT_BUCKET = "n_push"
const NOTIFICAION_WEBHOOK_RATE_LIMIT_BUCKET = "n_webhooks"

const NOTIFICATION_TEST_EMAIL_RATE_LIMIT_BUCKET = "n_test_mails"

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

		// Record metrics related to Notification Queue like size of queue and duration of pending notifications
		collectNotificationQueueMetrics()

		err = garbageCollectSentEvents()
		if err != nil {
			log.Error(err, "error garbage collecting sent notifications", 0)
		}

		err = garbageCollectOldPendingEvents()
		if err != nil {
			log.Error(err, "error garbage collecting old pending notifications", 0)
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

func garbageCollectSentEvents() error {
	rows, err := db.WriterDb.Exec(`DELETE FROM notification_queue WHERE sent < now() - INTERVAL '30 minutes'`)
	if err != nil {
		return fmt.Errorf("error deleting sent events from notification_queue %w", err)
	}

	rowsAffected, _ := rows.RowsAffected()

	log.Infof("deleted %v sent events from the notification_queue", rowsAffected)

	metrics.NotificationsDropped.WithLabelValues(string(Sent)).Add(float64(rowsAffected))

	return nil
}

func garbageCollectOldPendingEvents() error {
	rows, err := db.WriterDb.Exec(`DELETE FROM notification_queue WHERE created < now() - INTERVAL '1 hour'`)
	if err != nil {
		return fmt.Errorf("error deleting pending events from notification_queue %w", err)
	}

	rowsAffected, _ := rows.RowsAffected()

	log.Infof("deleted %v old pending events from the notification_queue", rowsAffected)

	metrics.NotificationsDropped.WithLabelValues(string(Pending)).Add(float64(rowsAffected))

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
		emailNotificationsPerDay := uint64(10)
		if err != nil {
			log.Error(err, "error getting user info", 0)
		} else {
			emailNotificationsPerDay = userInfo.PremiumPerks.EmailNotificationsPerDay
		}
		err = mail.SendMailRateLimited(n.Content, int64(emailNotificationsPerDay), NOTIFICAION_EMAIL_RATE_LIMIT_BUCKET)
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

	// webhooks have 5 seconds to respond
	client := &http.Client{Timeout: time.Second * 5}

	log.Infof("processing %v webhook notifications", len(notificationQueueItem))

	// use an error group to throttle webhook requests
	g := &errgroup.Group{}
	g.SetLimit(50) // issue at most 50 requests at a time
	for _, n := range notificationQueueItem {
		n := n
		_, err := db.CountSentMessage(NOTIFICAION_WEBHOOK_RATE_LIMIT_BUCKET, n.Content.UserId)
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

		g.Go(func() error {
			if n.Content.Webhook.Retries > 0 {
				time.Sleep(time.Duration(n.Content.Webhook.Retries) * time.Second)
			}
			resp, err := client.Post(n.Content.Webhook.Url, "application/json", reqBody)
			if err != nil {
				log.Warnf("error sending webhook request: %v", err)
				metrics.NotificationsSent.WithLabelValues("webhook", "error").Inc()
				return nil
			} else {
				metrics.NotificationsSent.WithLabelValues("webhook", resp.Status).Inc()
			}
			defer resp.Body.Close()

			_, err = db.WriterDb.Exec(`UPDATE notification_queue SET sent = now() WHERE id = $1`, n.Id)
			if err != nil {
				log.Error(err, "error updating notification_queue table", 0)
				return nil
			}

			if resp != nil && resp.StatusCode < 400 {
				// update retries counters in db based on end result
				if n.Content.Webhook.DashboardId == 0 && n.Content.Webhook.DashboardGroupId == 0 {
					_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = $1, last_sent = now() WHERE id = $2;`, n.Content.Webhook.Retries, n.Content.Webhook.ID)
				} else {
					_, err = db.WriterDb.Exec(`UPDATE users_val_dashboards_groups SET webhook_retries = $1, webhook_last_sent = now() WHERE id = $2 AND dashboard_id = $3;`, n.Content.Webhook.Retries, n.Content.Webhook.DashboardGroupId, n.Content.Webhook.DashboardId)
				}
				if err != nil {
					log.Warnf("failed to update retries counter to %v for webhook %v: %v", n.Content.Webhook.Retries, n.Content.Webhook.ID, err)
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

				if n.Content.Webhook.DashboardId == 0 && n.Content.Webhook.DashboardGroupId == 0 {
					_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = retries + 1, last_sent = now(), request = $2, response = $3 WHERE id = $1;`, n.Content.Webhook.ID, n.Content, errResp)
				} else {
					_, err = db.WriterDb.Exec(`UPDATE users_val_dashboards_groups SET webhook_retries = webhook_retries + 1, webhook_last_sent = now() WHERE id = $1 AND dashboard_id = $2;`, n.Content.Webhook.DashboardGroupId, n.Content.Webhook.DashboardId)
				}
				if err != nil {
					log.Error(err, "error updating users_webhooks table", 0)
					return nil
				}
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		log.Error(err, "error waiting for errgroup", 0)
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
	// webhooks have 5 seconds to respond
	client := &http.Client{Timeout: time.Second * 5}

	log.Infof("processing %v discord webhook notifications", len(notificationQueueItem))

	// use an error group to throttle webhook requests
	g := &errgroup.Group{}
	g.SetLimit(50) // issue at most 50 requests at a time
	for _, n := range notificationQueueItem {
		n := n
		_, err := db.CountSentMessage(NOTIFICAION_WEBHOOK_RATE_LIMIT_BUCKET, n.Content.UserId)
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

		err = json.NewEncoder(reqBody).Encode(n.Content.DiscordRequest)
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

		g.Go(func() error {
			if n.Content.Webhook.Retries > 0 {
				time.Sleep(time.Duration(n.Content.Webhook.Retries) * time.Second)
			}
			resp, err := client.Post(n.Content.Webhook.Url, "application/json", reqBody)
			if err != nil {
				log.Warnf("error sending discord webhook request: %v", err)
				metrics.NotificationsSent.WithLabelValues("webhook_discord", "error").Inc()
				return nil
			} else {
				metrics.NotificationsSent.WithLabelValues("webhook_discord", resp.Status).Inc()
			}
			defer resp.Body.Close()

			_, err = db.WriterDb.Exec(`UPDATE notification_queue SET sent = now() WHERE id = $1`, n.Id)
			if err != nil {
				log.Error(err, "error updating notification_queue table", 0)
				return nil
			}

			if resp != nil && resp.StatusCode < 400 {
				// update retries counters in db based on end result
				if n.Content.Webhook.DashboardId == 0 && n.Content.Webhook.DashboardGroupId == 0 {
					_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = $1, last_sent = now() WHERE id = $2;`, n.Content.Webhook.Retries, n.Content.Webhook.ID)
				} else {
					_, err = db.WriterDb.Exec(`UPDATE users_val_dashboards_groups SET webhook_retries = $1, webhook_last_sent = now() WHERE id = $2 AND dashboard_id = $3;`, n.Content.Webhook.Retries, n.Content.Webhook.DashboardGroupId, n.Content.Webhook.DashboardId)
				}
				if err != nil {
					log.Warnf("failed to update retries counter to %v for webhook %v: %v", n.Content.Webhook.Retries, n.Content.Webhook.ID, err)
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

				if n.Content.Webhook.DashboardId == 0 && n.Content.Webhook.DashboardGroupId == 0 {
					_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = retries + 1, last_sent = now(), request = $2, response = $3 WHERE id = $1;`, n.Content.Webhook.ID, n.Content, errResp)
				} else {
					_, err = db.WriterDb.Exec(`UPDATE users_val_dashboards_groups SET webhook_retries = webhook_retries + 1, webhook_last_sent = now() WHERE id = $1 AND dashboard_id = $2;`, n.Content.Webhook.DashboardGroupId, n.Content.Webhook.DashboardId)
				}
				if err != nil {
					log.Error(err, "error updating users_webhooks table", 0)
					return nil
				}
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		log.Error(err, "error waiting for errgroup", 0)
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
	err = mail.SendMailRateLimited(content, 10, NOTIFICATION_TEST_EMAIL_RATE_LIMIT_BUCKET)
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
	if count > 100 {
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

type Notification struct {
	Id      *uint64    `db:"id"`
	Created *time.Time `db:"created"`
	Sent    *time.Time `db:"sent"`
	Channel string     `db:"channel"`
	Content string     `db:"content"`
}

type NotificationStatus string

const (
	Sent    NotificationStatus = "sent"
	Pending NotificationStatus = "pending"

	// Metrics label for EventType where the event could not be mapped
	UnknownEvent string = "unknown_event"
)

/**
 * Get all notifications which were marked as sent.
 */
func GetSentNotifications() ([]Notification, error) {
	notificationRecords := []Notification{}

	err := db.ReaderDb.Select(&notificationRecords,
		`SELECT id, created, sent, channel, content 
		   FROM notification_queue
		  WHERE sent IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("error querying sent notifications: %w", err)
	}

	return notificationRecords, nil
}

/**
 * Get all notifications which have not yet been sent.
 */
func GetPendingNotifications() ([]Notification, error) {
	notificationRecords := []Notification{}

	err := db.ReaderDb.Select(&notificationRecords,
		`SELECT id, created, sent, channel, content 
		   FROM notification_queue
		  WHERE sent IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("error querying pending notifications: %w", err)
	}

	return notificationRecords, nil
}

/**
 * Collects metrics for the Notification system, specifically about its queue.
 * Provides metrics related to how many notifications are pending and sent, and of what event type they are.
 * Also provides metrics related to how long the notifications have been in the queue
 */
func collectNotificationQueueMetrics() {
	sentNotifications, err := GetSentNotifications()
	if err != nil {
		log.Error(err, "Error retrieving sent notifications. Will skip sending metrics", 0)
		return // Don't return an error, we don't want to disrupt actual notification sending simply because we couldn't record metrics
	}
	pendingNotifications, err := GetPendingNotifications()
	if err != nil {
		log.Error(err, "Error retrieving pending notifications. Will skip sending metrics", 0)
		return // Don't return an error, we don't want to disrupt actual notification sending simply because we couldn't record metrics
	}

	now := time.Now() // Checking the time once so that it is consistent across all metrics for this collection attempt

	// Record for each sent notification how long it took to send. Honestly, can probably remove this later since this metric can be sent once when the notification itself is delivered.
	for _, notification := range sentNotifications {
		eventType := GetEventLabelForNotification(notification)

		// Record the amount of time records that were sent (and that still exist in the queue) took to sent
		metrics.NotificationsQueueSentTime.WithLabelValues(notification.Channel, eventType).Observe(GetTimeDiffMilliseconds(*notification.Sent, *notification.Created))
	}

	// Record for each pending notification how long it has been in the queue
	for _, notification := range pendingNotifications {
		eventType := GetEventLabelForNotification(notification)

		// Record the amount of time these records have been waiting to been sent
		metrics.NotificationsQueuePendingTime.WithLabelValues(notification.Channel, eventType).Observe(GetTimeDiffMilliseconds(*notification.Created, now))
	}

	// Count number of pending notifications in the queue by event type
	eventTypeCount := CountByEventType(pendingNotifications)
	for eventType, numNotifications := range eventTypeCount {
		metrics.NotificationsQueueEventSize.WithLabelValues(eventType, string(Pending)).Set(float64(numNotifications))
	}

	// Count number of sent notifications in the queue by event type
	eventTypeCount = CountByEventType(sentNotifications)
	for eventType, numNotifications := range eventTypeCount {
		metrics.NotificationsQueueEventSize.WithLabelValues(eventType, string(Sent)).Set(float64(numNotifications))
	}

	// Count number of pending notifications in the queue by channel
	channelCount := CountByChannel(pendingNotifications)
	for channelType, numNotifications := range channelCount {
		metrics.NotificationsQueueChannelSize.WithLabelValues(channelType, string(Pending)).Set(float64(numNotifications))
	}

	// Count number of sent notifications in the queue by channel
	channelCount = CountByChannel(sentNotifications)
	for channelType, numNotifications := range channelCount {
		metrics.NotificationsQueueChannelSize.WithLabelValues(channelType, string(Sent)).Set(float64(numNotifications))
	}
}

/**
 * Simple wrapper that enables submitting metrics for notifications with unknown event names.
 */
func GetEventLabelForNotification(notification Notification) string {
	eventName, err := ExtractEventNameFromNotification(notification)
	if err != nil {
		return UnknownEvent
	}

	return string(*eventName)
}

/**
 * Because we don't record the event type when recording notifications, we have to do some work to extract them from the
 * notification message that is eventually sent to the user.
 */
func ExtractEventNameFromNotification(notification Notification) (*types.EventName, error) {
	for eventName, eventDescription := range types.EventLabel {
		if strings.Contains(notification.Content, eventDescription) {
			return &eventName, nil
		}
	}

	// Also grab legacy labels, unfortunately some systems still submit these.
	for eventName, eventDescription := range types.LegacyEventLabel {
		if strings.Contains(notification.Content, eventDescription) {
			return &eventName, nil
		}
	}

	return nil, fmt.Errorf("no EventName found for notification %d matching any event descriptions", notification.Id)
}

/**
 * Given a collection of notifications, count the number of notifications with each distinct event type.
 */
func CountByEventType(notifications []Notification) map[string]int {
	eventTypeCountMap := make(map[string]int, len(types.EventLabel)+1) // +1 to account for the "unknown" event type
	// Initialize the map with all EventLabel types, with the Count set to 0
	// Must be pre-initialized, because we still want to submit metrics for 0-count EventTypes, so they must exist in this map.
	for eventType := range types.EventLabel {
		eventTypeCountMap[string(eventType)] = 0
	}
	eventTypeCountMap[UnknownEvent] = 0 // include unknown, which indicates an EventType which we couldn't parse from the Notification's content field

	// Now iterate over the list of events, and increment the value in the map
	for _, notification := range notifications {
		eventType := GetEventLabelForNotification(notification)
		eventTypeCountMap[eventType] = eventTypeCountMap[eventType] + 1
	}
	return eventTypeCountMap
}

/**
 * Given a collection of notifications, count the number of notifications with each distinct channel type.
 */
func CountByChannel(notifications []Notification) map[string]int {
	channelCountMap := make(map[string]int, len(types.NotificationChannels))
	// Initialize the map with the Channel types, with the Count set to 0.
	// Must be pre-initialized, because we still want to submit metrics for 0-count Channels, so they must exist in this map.
	for _, channelType := range types.NotificationChannels {
		channelCountMap[string(channelType)] = 0
	}

	// Now iterate over the list of notifications, and increment the value for each channel
	for _, notification := range notifications {
		channelCountMap[notification.Channel] = channelCountMap[notification.Channel] + 1
	}
	return channelCountMap
}

/**
 * Returns the amount of milliseconds between two timestamps. Always returns a positive
 * duration, so you don't have to worry about date ordering
 */
func GetTimeDiffMilliseconds(time1 time.Time, time2 time.Time) float64 {
	duration := time1.Sub(time2)
	return math.Abs(float64(duration.Milliseconds()))
}
