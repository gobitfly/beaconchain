package notification

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/gob"
	"fmt"
	"html/template"
	"maps"
	"slices"
	"strings"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func queueNotifications(epoch uint64, notificationsByUserID types.NotificationsPerUserId) error {
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer utils.Rollback(tx)

	err = QueueEmailNotifications(epoch, notificationsByUserID, tx)
	if err != nil {
		return fmt.Errorf("error queuing email notifications: %w", err)
	}

	err = QueuePushNotification(epoch, notificationsByUserID, tx)
	if err != nil {
		return fmt.Errorf("error queuing push notifications: %w", err)
	}

	err = QueueWebhookNotifications(notificationsByUserID, tx)
	if err != nil {
		return fmt.Errorf("error queuing webhook notifications: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	err = ExportNotificationHistory(epoch, notificationsByUserID)
	if err != nil {
		return fmt.Errorf("error exporting notification history: %w", err)
	}

	subByEpoch := map[uint64][]uint64{}
	for _, notificationsPerDashboard := range notificationsByUserID {
		for _, notificationsPerGroup := range notificationsPerDashboard {
			for _, events := range notificationsPerGroup {
				for _, notifications := range events {
					for _, n := range notifications {
						e := n.GetEpoch()
						if _, exists := subByEpoch[e]; !exists {
							subByEpoch[e] = []uint64{n.GetSubscriptionID()}
						} else {
							subByEpoch[e] = append(subByEpoch[e], n.GetSubscriptionID())
						}
					}
				}
			}
		}
	}

	// obsolete as notifications are anyway sent on a per-epoch basis
	for epoch, subIDs := range subByEpoch {
		// update that we've queued the subscription (last sent rather means last queued)
		err := db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch)
		if err != nil {
			log.Error(err, "error updating sent-time of sent notifications", 0)
			metrics.Errors.WithLabelValues("notifications_updating_sent_time").Inc()
		}
	}
	// update internal state of subscriptions
	// stateToSub := make(map[string]map[uint64]bool, 0)

	// for _, notificationMap := range notificationsByUserID { // _ => user
	// 	for _, notifications := range notificationMap { // _ => eventname
	// 		for _, notification := range notifications { // _ => index
	// 			state := notification.GetLatestState()
	// 			if state == "" {
	// 				continue
	// 			}
	// 			if _, exists := stateToSub[state]; !exists {
	// 				stateToSub[state] = make(map[uint64]bool, 0)
	// 			}
	// 			if _, exists := stateToSub[state][notification.GetSubscriptionID()]; !exists {
	// 				stateToSub[state][notification.GetSubscriptionID()] = true
	// 			}
	// 		}
	// 	}
	// }

	// no need to batch here as the internal state will become obsolete
	// for state, subs := range stateToSub {
	// 	subArray := make([]int64, 0)
	// 	for subID := range subs {
	// 		subArray = append(subArray, int64(subID))
	// 	}
	// 	_, err := db.FrontendWriterDB.Exec(`UPDATE users_subscriptions SET internal_state = $1 WHERE id = ANY($2)`, state, pq.Int64Array(subArray))
	// 	if err != nil {
	// 		log.Error(err, "error updating internal state of notifications", 0)
	// 	}
	// }
	return nil
}

func ExportNotificationHistory(epoch uint64, notificationsByUserID types.NotificationsPerUserId) error {
	dashboardNotificationHistoryInsertStmt, err := db.WriterDb.Preparex(`
		INSERT INTO users_val_dashboards_notifications_history 
		(user_id, dashboard_id, group_id, epoch, event_type, event_count, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement for dashboard notifications history: %w", err)
	}
	defer utils.ClosePreparedStatement(dashboardNotificationHistoryInsertStmt)

	machineNotificationHistoryInsertStmt, err := db.FrontendWriterDB.Preparex(`
		INSERT INTO machine_notifications_history 
		(user_id, epoch, machine_id, machine_name, event_type, event_threshold)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement for machine notifications history: %w", err)
	}
	defer utils.ClosePreparedStatement(machineNotificationHistoryInsertStmt)

	clientNotificationHistoryInsertStmt, err := db.FrontendWriterDB.Preparex(`
		INSERT INTO client_notifications_history 
		(user_id, epoch, client, client_version, client_url)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement for client notifications history: %w", err)
	}
	defer utils.ClosePreparedStatement(clientNotificationHistoryInsertStmt)

	networktNotificationHistoryInsertStmt, err := db.FrontendWriterDB.Preparex(`
		INSERT INTO network_notifications_history 
		(user_id, epoch, network, event_type, event_threshold)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement for client notifications history: %w", err)
	}
	defer utils.ClosePreparedStatement(networktNotificationHistoryInsertStmt)

	for userID, notificationsPerDashboard := range notificationsByUserID {
		for dashboardID, notificationsPerGroup := range notificationsPerDashboard {
			for group, notifications := range notificationsPerGroup {
				for eventName, notifications := range notifications {
					// handle all dashboard related notifications
					if eventName != types.NetworkLivenessIncreasedEventName && !types.IsUserIndexed(eventName) && !types.IsMachineNotification(eventName) {
						details, err := GetNotificationDetails(notifications)
						if err != nil {
							return fmt.Errorf("error getting notification details: %w", err)
						}
						_, err = dashboardNotificationHistoryInsertStmt.Exec(
							userID,
							dashboardID,
							group,
							epoch,
							eventName,
							len(notifications),
							details,
						)
						if err != nil {
							return fmt.Errorf("error inserting into dashboard notifications history: %w", err)
						}
					} else if types.IsMachineNotification(eventName) { // handle machine monitoring related events
						for _, n := range notifications {
							nTyped, ok := n.(*MonitorMachineNotification)
							if !ok {
								return fmt.Errorf("error casting machine notification: %w", err)
							}
							_, err := machineNotificationHistoryInsertStmt.Exec(
								userID,
								epoch,
								0,
								nTyped.MachineName,
								eventName,
								nTyped.EventThreshold,
							)
							if err != nil {
								return fmt.Errorf("error inserting into machine notifications history: %w", err)
							}
						}
					} else if eventName == types.EthClientUpdateEventName { // handle client update events
						for _, n := range notifications {
							nTyped, ok := n.(*EthClientNotification)
							if !ok {
								return fmt.Errorf("error casting client update notification: %w", err)
							}
							_, err := clientNotificationHistoryInsertStmt.Exec(
								userID,
								epoch,
								nTyped.EthClient,
								"",
								"",
							)
							if err != nil {
								return fmt.Errorf("error inserting into client notifications history: %w", err)
							}
						}
					} else if eventName == types.NetworkLivenessIncreasedEventName { // handle network liveness increased events
						for range notifications {
							_, err := networktNotificationHistoryInsertStmt.Exec(
								userID,
								epoch,
								utils.Config.Chain.Name,
								eventName,
								0,
							)
							if err != nil {
								return fmt.Errorf("error inserting into network notifications history: %w", err)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func GetNotificationDetails(notificationsPerEventFilter types.NotificationsPerEventFilter) ([]byte, error) {
	// get the notifications as array
	notifications := make([]types.Notification, 0, len(notificationsPerEventFilter))
	for _, ns := range notificationsPerEventFilter {
		ns.SetEventFilter("") // zero out the event filter as it is not needed in the details
		notifications = append(notifications, ns)
	}
	// gob encode and gzip compress the notifications
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(gz)
	err := enc.Encode(notifications)
	if err != nil {
		return nil, fmt.Errorf("error encoding notifications: %w", err)
	}
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("error compressing notifications: %w", err)
	}
	return buf.Bytes(), nil
}

func RenderEmailsForUserEvents(epoch uint64, notificationsByUserID types.NotificationsPerUserId) (emails []types.TransitEmailContent, err error) {
	emails = make([]types.TransitEmailContent, 0, 50)

	createdTs := time.Now()

	userIDs := slices.Collect(maps.Keys(notificationsByUserID))

	emailsByUserID, err := GetUserEmailsByIds(userIDs)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_get_user_mail_by_id").Inc()
		return nil, fmt.Errorf("error when sending email-notifications: could not get emails: %w", err)
	}

	for userID, notificationsPerDashboard := range notificationsByUserID {
		userEmail, exists := emailsByUserID[userID]
		if !exists {
			log.WarnWithStackTrace(nil, "email notification skipping user", 0, log.Fields{"user_id": userID})
			// we don't need this metrics as users can now deactivate email notifications and it would increment the counter
			// metrics.Errors.WithLabelValues("notifications_mail_not_found").Inc()
			continue
		}
		attachments := []types.EmailAttachment{}

		var msg types.Email

		if utils.Config.Chain.Name != "mainnet" {
			//nolint:gosec // this is a static string
			msg.Body += template.HTML(fmt.Sprintf("<b>Notice: This email contains notifications for the %s network!</b><br>", utils.Config.Chain.Name))
		}

		subject := ""
		notificationTypesMap := make(map[types.EventName]int)
		uniqueNotificationTypes := []types.EventName{}

		bodyDetails := template.HTML("")

		for _, event := range types.EventSortOrder {
			for _, notificationsPerGroup := range notificationsPerDashboard {
				for _, userNotifications := range notificationsPerGroup {
					ns, ok := userNotifications[event]
					if !ok { // nothing to do for this event type
						continue
					}

					if len(bodyDetails) > 0 {
						bodyDetails += "<br>"
					}
					//nolint:gosec // this is a static string
					bodyDetails += template.HTML(fmt.Sprintf("<u>%s</u><br>", types.EventLabel[event]))
					i := 0
					for _, n := range ns {
						// Find all unique notification titles for the subject
						if _, ok := notificationTypesMap[event]; !ok {
							uniqueNotificationTypes = append(uniqueNotificationTypes, event)
						}
						notificationTypesMap[event]++

						if i <= 10 {
							if event != types.SyncCommitteeSoonEventName {
								// SyncCommitteeSoon notifications are summed up in getEventInfo for all validators
								//nolint:gosec // this is a static string
								bodyDetails += template.HTML(fmt.Sprintf("%s<br>", n.GetInfo(types.NotifciationFormatHtml)))
							}

							if att := n.GetEmailAttachment(); att != nil {
								attachments = append(attachments, *att)
							}
						}

						metrics.NotificationsQueued.WithLabelValues("email", string(event)).Inc()
						i++

						if i == 11 {
							//nolint:gosec // this is a static string
							bodyDetails += template.HTML(fmt.Sprintf("... and %d more notifications<br>", len(ns)-i))
							continue
						}
					}

					eventInfo := getEventInfo(event, types.NotifciationFormatHtml, ns)
					if eventInfo != "" {
						//nolint:gosec // this is a static string
						bodyDetails += template.HTML(fmt.Sprintf("%s<br>", eventInfo))
					}
				}
			}
		}

		//nolint:gosec // this is a static string
		bodySummary := template.HTML(fmt.Sprintf("<h5>Summary for epoch %d:</h5>", epoch))
		for _, event := range types.EventSortOrder {
			count, ok := notificationTypesMap[event]
			if !ok {
				continue
			}
			if len(bodySummary) > 0 {
				bodySummary += "<br>"
			}
			plural := ""
			if count > 1 {
				plural = "s"
			}
			switch event {
			case types.RocketpoolCollateralMaxReachedEventName, types.RocketpoolCollateralMinReachedEventName:
				//nolint:gosec // this is a static string
				bodySummary += template.HTML(fmt.Sprintf("%s: %d node%s", types.EventLabel[event], count, plural))
			case types.TaxReportEventName, types.NetworkLivenessIncreasedEventName:
				//nolint:gosec // this is a static string
				bodySummary += template.HTML(fmt.Sprintf("%s: %d event%s", types.EventLabel[event], count, plural))
			case types.EthClientUpdateEventName:
				//nolint:gosec // this is a static string
				bodySummary += template.HTML(fmt.Sprintf("%s: %d client%s", types.EventLabel[event], count, plural))
			default:
				//nolint:gosec // this is a static string
				bodySummary += template.HTML(fmt.Sprintf("%s: %d Validator%s", types.EventLabel[event], count, plural))
			}
		}
		msg.Body += bodySummary
		msg.Body += template.HTML("<br><br><h5>Details:</h5>")
		msg.Body += bodyDetails

		if len(uniqueNotificationTypes) > 2 {
			subject = fmt.Sprintf("%s: %s,... and %d other notifications", utils.Config.Frontend.SiteDomain, types.EventLabel[uniqueNotificationTypes[0]], len(uniqueNotificationTypes)-1)
		} else if len(uniqueNotificationTypes) == 2 {
			subject = fmt.Sprintf("%s: %s and %s", utils.Config.Frontend.SiteDomain, types.EventLabel[uniqueNotificationTypes[0]], types.EventLabel[uniqueNotificationTypes[1]])
		} else if len(uniqueNotificationTypes) == 1 {
			subject = fmt.Sprintf("%s: %s", utils.Config.Frontend.SiteDomain, types.EventLabel[uniqueNotificationTypes[0]])
		}
		//nolint:gosec // this is a static string
		msg.SubscriptionManageURL = template.HTML(fmt.Sprintf(`<a href="%v" style="color: white" onMouseOver="this.style.color='#F5B498'" onMouseOut="this.style.color='#FFFFFF'">Manage</a>`, "https://"+utils.Config.Frontend.SiteDomain+"/user/notifications"))

		emails = append(emails, types.TransitEmailContent{
			Address:     userEmail,
			Subject:     subject,
			Email:       msg,
			Attachments: attachments,
			CreatedTs:   createdTs,
		})
	}
	return emails, nil
}

func QueueEmailNotifications(epoch uint64, notificationsByUserID types.NotificationsPerUserId, tx *sqlx.Tx) error {
	// for emails multiple notifications will be rendered to one email per user for each run
	emails, err := RenderEmailsForUserEvents(epoch, notificationsByUserID)
	if err != nil {
		return fmt.Errorf("error rendering emails: %w", err)
	}

	// now batch insert the emails in one go
	log.Infof("queueing %v email notifications", len(emails))
	type insertData struct {
		Content types.TransitEmailContent `db:"content"`
	}

	insertRows := make([]insertData, 0, len(emails))
	for _, email := range emails {
		insertRows = append(insertRows, insertData{
			Content: email,
		})
	}

	_, err = tx.NamedExec(`INSERT INTO notification_queue (created, channel, content) VALUES (NOW(), 'email', :content)`, insertRows)
	if err != nil {
		log.Error(err, "error writing transit email to db", 0)
	}
	return nil
}

func RenderPushMessagesForUserEvents(epoch uint64, notificationsByUserID types.NotificationsPerUserId) ([]types.TransitPushContent, error) {
	pushMessages := make([]types.TransitPushContent, 0, 50)

	userIDs := slices.Collect(maps.Keys(notificationsByUserID))

	tokensByUserID, err := GetUserPushTokenByIds(userIDs)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_send_push_notifications").Inc()
		return nil, fmt.Errorf("error when sending push-notifications: could not get tokens: %w", err)
	}

	for userID, notificationsPerDashboard := range notificationsByUserID {
		userTokens, exists := tokensByUserID[userID]
		if !exists {
			continue
		}
		for dashboardId, notficationsPerGroup := range notificationsPerDashboard {
			for groupdId, userNotifications := range notficationsPerGroup {
				log.Infof("generating push notification for user %d & dashboard %d and group %d", userID, dashboardId, groupdId)

				notificationTypesMap := make(map[types.EventName][]string)

				for _, event := range types.EventSortOrder {
					ns, ok := userNotifications[event]
					if !ok { // nothing to do for this event type
						continue
					}
					if _, ok := notificationTypesMap[event]; !ok {
						notificationTypesMap[event] = make([]string, 0)
					}
					for _, n := range ns {
						notificationTypesMap[event] = append(notificationTypesMap[event], n.GetEntitiyId())
					}
					metrics.NotificationsQueued.WithLabelValues("push", string(event)).Inc()
				}

				bodySummary := ""
				for _, event := range types.EventSortOrder {
					events := notificationTypesMap[event]
					if len(events) == 0 {
						continue
					}
					count := len(events)
					if len(bodySummary) > 0 {
						bodySummary += "\n"
					}
					plural := ""
					if count > 1 {
						plural = "s"
					}
					switch event {
					case types.RocketpoolCollateralMaxReachedEventName, types.RocketpoolCollateralMinReachedEventName:
						bodySummary += fmt.Sprintf("%s: %d node%s", types.EventLabel[event], count, plural)
					case types.TaxReportEventName, types.NetworkLivenessIncreasedEventName:
						bodySummary += fmt.Sprintf("%s: %d event%s", types.EventLabel[event], count, plural)
					case types.EthClientUpdateEventName:
						bodySummary += fmt.Sprintf("%s: %d client%s", types.EventLabel[event], count, plural)
					case types.MonitoringMachineCpuLoadEventName, types.MonitoringMachineMemoryUsageEventName, types.MonitoringMachineDiskAlmostFullEventName, types.MonitoringMachineOfflineEventName:
						bodySummary += fmt.Sprintf("%s: %d machine%s", types.EventLabel[event], count, plural)
					default:
						bodySummary += fmt.Sprintf("%s: %d validator%s", types.EventLabel[event], count, plural)
					}
					truncated := ""
					if len(events) > 3 {
						truncated = ",..."
						events = events[:3]
					}
					bodySummary += fmt.Sprintf(" (%s%s)", strings.Join(events, ","), truncated)
				}

				if len(bodySummary) > 1000 { // cap the notification body to 1000 characters (firebase limit)
					bodySummary = bodySummary[:1000]
				}
				for _, userToken := range userTokens {
					message := new(messaging.Message)
					message.Token = userToken
					message.APNS = new(messaging.APNSConfig)
					message.APNS.Payload = new(messaging.APNSPayload)
					message.APNS.Payload.Aps = new(messaging.Aps)
					message.APNS.Payload.Aps.Sound = "default"

					notification := new(messaging.Notification)
					notification.Title = fmt.Sprintf("%sInfo for epoch %d", getNetwork(), epoch)
					notification.Body = bodySummary
					message.Notification = notification
					message.Data = map[string]string{
						"epoch": fmt.Sprintf("%d", epoch),
					}
					transitPushContent := types.TransitPushContent{
						Messages: []*messaging.Message{message},
					}

					pushMessages = append(pushMessages, transitPushContent)
				}
			}
		}
	}

	return pushMessages, nil
}

func QueuePushNotification(epoch uint64, notificationsByUserID types.NotificationsPerUserId, tx *sqlx.Tx) error {
	pushMessages, err := RenderPushMessagesForUserEvents(epoch, notificationsByUserID)
	if err != nil {
		return fmt.Errorf("error rendering push messages: %w", err)
	}

	// now batch insert the push messages in one go
	log.Infof("queueing %v push notifications", len(pushMessages))
	type insertData struct {
		Content types.TransitPushContent `db:"content"`
	}

	insertRows := make([]insertData, 0, len(pushMessages))
	for _, pushMessage := range pushMessages {
		insertRows = append(insertRows, insertData{
			Content: pushMessage,
		})
	}

	_, err = tx.NamedExec(`INSERT INTO notification_queue (created, channel, content) VALUES (NOW(), 'push', :content)`, insertRows)
	if err != nil {
		return fmt.Errorf("error writing transit push to db: %w", err)
	}
	return nil
}

func QueueWebhookNotifications(notificationsByUserID types.NotificationsPerUserId, tx *sqlx.Tx) error {
	for userID, userNotifications := range notificationsByUserID {
		var webhooks []types.UserWebhook
		err := db.FrontendWriterDB.Select(&webhooks, `
			SELECT
				id,
				user_id,
				url,
				retries,
				event_names,
				last_sent,
				destination
			FROM
				users_webhooks
			WHERE
				user_id = $1 AND user_id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)
		`, userID, types.WebhookNotificationChannel)
		// continue if the user does not have a webhook
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("error quering users_webhooks, err: %w", err)
		}
		// webhook => [] notifications
		discordNotifMap := make(map[uint64][]types.TransitDiscordContent)
		notifs := make([]types.TransitWebhook, 0)
		// send the notifications to each registered webhook
		for _, w := range webhooks {
			for dashboardId, notificationsPerDashboard := range userNotifications {
				if dashboardId != 0 { // disable webhooks for dashboard notifications for now
					continue
				}
				for _, notificationsPerGroup := range notificationsPerDashboard {
					for event, notifications := range notificationsPerGroup {
						// check if the webhook is subscribed to the type of event
						eventSubscribed := slices.Contains(w.EventNames, string(event))

						if eventSubscribed {
							if len(notifications) > 0 {
								// reset Retries
								if w.Retries > 5 && w.LastSent.Valid && w.LastSent.Time.Add(time.Hour).Before(time.Now()) {
									_, err = db.FrontendWriterDB.Exec(`UPDATE users_webhooks SET retries = 0 WHERE id = $1;`, w.ID)
									if err != nil {
										log.Error(err, "error updating users_webhooks table; setting retries to zero", 0)
										continue
									}
								} else if w.Retries > 5 && !w.LastSent.Valid {
									log.Warnf("webhook '%v' has more than 5 retries and does not have a valid last_sent timestamp", w.Url)
									continue
								}

								if w.Retries >= 5 {
									// early return
									continue
								}
							}

							for _, n := range notifications {
								if w.Destination.Valid && w.Destination.String == "webhook_discord" {
									if _, exists := discordNotifMap[w.ID]; !exists {
										discordNotifMap[w.ID] = make([]types.TransitDiscordContent, 0)
									}
									l_notifs := len(discordNotifMap[w.ID])
									if l_notifs == 0 || len(discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds) >= 10 {
										discordNotifMap[w.ID] = append(discordNotifMap[w.ID], types.TransitDiscordContent{
											Webhook: w,
											DiscordRequest: types.DiscordReq{
												Username: utils.Config.Frontend.SiteDomain,
											},
										})
										l_notifs++
									}

									fields := []types.DiscordEmbedField{
										{
											Name:   "Epoch",
											Value:  fmt.Sprintf("[%[1]v](https://%[2]s/%[1]v)", n.GetEpoch(), utils.Config.Frontend.SiteDomain+"/epoch"),
											Inline: false,
										},
									}

									if strings.HasPrefix(string(n.GetEventName()), "monitoring") || n.GetEventName() == types.EthClientUpdateEventName || n.GetEventName() == types.RocketpoolCollateralMaxReachedEventName || n.GetEventName() == types.RocketpoolCollateralMinReachedEventName {
										fields = append(fields,
											types.DiscordEmbedField{
												Name:   "Target",
												Value:  fmt.Sprintf("%v", n.GetEventFilter()),
												Inline: false,
											})
									}
									discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds = append(discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds, types.DiscordEmbed{
										Type:        "rich",
										Color:       "16745472",
										Description: n.GetLegacyInfo(),
										Title:       n.GetLegacyTitle(),
										Fields:      fields,
									})
								} else {
									notifs = append(notifs, types.TransitWebhook{
										Channel: w.Destination.String,
										Content: types.TransitWebhookContent{
											Webhook: w,
											Event: types.WebhookEvent{
												Network:     utils.GetNetwork(),
												Name:        string(n.GetEventName()),
												Title:       n.GetLegacyTitle(),
												Description: n.GetLegacyInfo(),
												Epoch:       n.GetEpoch(),
												Target:      n.GetEventFilter(),
											},
										},
									})
								}
							}
						}
					}
				}
			}
		}
		// process notifs
		for _, n := range notifs {
			_, err = tx.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES (now(), $1, $2);`, n.Channel, n.Content)
			if err != nil {
				log.Error(err, "error inserting into webhooks_queue", 0)
			} else {
				metrics.NotificationsQueued.WithLabelValues(n.Channel, n.Content.Event.Name).Inc()
			}
		}
		// process discord notifs
		for _, dNotifs := range discordNotifMap {
			for _, n := range dNotifs {
				_, err = tx.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES (now(), 'webhook_discord', $1);`, n)
				if err != nil {
					log.Error(err, "error inserting into webhooks_queue (discord)", 0)
					continue
				} else {
					metrics.NotificationsQueued.WithLabelValues("webhook_discord", "multi").Inc()
				}
			}
		}
	}
	return nil
}

func getNetwork() string {
	domainParts := strings.Split(utils.Config.Frontend.SiteDomain, ".")
	if len(domainParts) >= 3 {
		return fmt.Sprintf("%s: ", cases.Title(language.English).String(domainParts[0]))
	}
	return ""
}

func getEventInfo(event types.EventName, format types.NotificationFormat, ns map[types.EventFilter]types.Notification) string {
	switch event {
	case types.SyncCommitteeSoonEventName:
		return getSyncCommitteeSoonInfo(format, ns)
	case "validator_balance_decreased":
		return "<br>You will not receive any further balance decrease mails for these validators until the balance of a validator is increasing again."
	}

	return ""
}