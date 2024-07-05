package notification

import (
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
)

// Retrieves all subscription for a given event filter
// Map key corresponds to the event filter which can be
// a validator pubkey or an eth1 address (for RPL notifications)
// or a list of validators for the tax report notifications
func GetSubsForEventFilter(eventName types.EventName) (map[string][]types.Subscription, error) {
	var subs []types.Subscription
	subQuery := `
		SELECT id, user_id, event_filter, last_sent_epoch, created_epoch, event_threshold, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash, internal_state from users_subscriptions where event_name = $1
		`

	subMap := make(map[string][]types.Subscription, 0)
	err := db.FrontendWriterDB.Select(&subs, subQuery, utils.GetNetwork()+":"+string(eventName))
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		if _, ok := subMap[sub.EventFilter]; !ok {
			subMap[sub.EventFilter] = make([]types.Subscription, 0)
		}
		subMap[sub.EventFilter] = append(subMap[sub.EventFilter], types.Subscription{
			UserID:         sub.UserID,
			ID:             sub.ID,
			LastEpoch:      sub.LastEpoch,
			EventFilter:    sub.EventFilter,
			CreatedEpoch:   sub.CreatedEpoch,
			EventThreshold: sub.EventThreshold,
			State:          sub.State,
		})
	}
	return subMap, nil
}

func GetUserPushTokenByIds(ids []types.UserId) (map[types.UserId][]string, error) {
	pushByID := map[types.UserId][]string{}
	if len(ids) == 0 {
		return pushByID, nil
	}
	var rows []struct {
		ID    types.UserId `db:"user_id"`
		Token string       `db:"notification_token"`
	}

	err := db.FrontendWriterDB.Select(&rows, "SELECT DISTINCT ON (user_id, notification_token) user_id, notification_token FROM users_devices WHERE (user_id = ANY($1) AND user_id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)) AND notify_enabled = true AND active = true AND notification_token IS NOT NULL AND LENGTH(notification_token) > 20 ORDER BY user_id, notification_token, id DESC", pq.Array(ids), types.PushNotificationChannel)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		val, ok := pushByID[r.ID]
		if ok {
			pushByID[r.ID] = append(val, r.Token)
		} else {
			pushByID[r.ID] = []string{r.Token}
		}
	}

	return pushByID, nil
}

// GetUserEmailsByIds returns the emails of users.
func GetUserEmailsByIds(ids []types.UserId) (map[types.UserId]string, error) {
	mailsByID := map[types.UserId]string{}
	if len(ids) == 0 {
		return mailsByID, nil
	}
	var rows []struct {
		ID    types.UserId `db:"id"`
		Email string       `db:"email"`
	}
	//
	err := db.FrontendWriterDB.Select(&rows, "SELECT id, email FROM users WHERE id = ANY($1) AND id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)", pq.Array(ids), types.EmailNotificationChannel)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		mailsByID[r.ID] = r.Email
	}
	return mailsByID, nil
}
