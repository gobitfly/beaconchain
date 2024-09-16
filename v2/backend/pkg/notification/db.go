package notification

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
)

// Retrieves all subscription for a given event filter
// Map key corresponds to the event filter which can be
// a validator pubkey or an eth1 address (for RPL notifications)
// or a list of validators for the tax report notifications
// or a machine name for machine notifications or a eth client name for ethereum client update notifications
// optionally it is possible to set a filter on the last sent ts and the event filter
// fields
func GetSubsForEventFilter(eventName types.EventName, lastSentFilter string, lastSentFilterArgs []interface{}, eventFilters []string) (map[string][]types.Subscription, error) {
	var subs []types.Subscription

	// subQuery := `
	// 	SELECT
	// 		id,
	// 		user_id,
	// 		event_filter,
	// 		last_sent_epoch,
	// 		created_epoch,
	// 		event_threshold,
	// 		ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash,
	// 		internal_state
	// 	from users_subscriptions
	// 	where event_name = $1
	// 	`

	ds := goqu.Dialect("postgres").From("users_subscriptions").Select(
		goqu.C("id"),
		goqu.C("user_id"),
		goqu.C("event_filter"),
		goqu.C("last_sent_epoch"),
		goqu.C("created_epoch"),
		goqu.C("event_threshold"),
		goqu.C("event_name"),
	).Where(goqu.L("(event_name = ? AND user_id <> 0)", utils.GetNetwork()+":"+string(eventName)))

	if lastSentFilter != "" {
		if len(lastSentFilterArgs) > 0 {
			ds = ds.Where(goqu.L(lastSentFilter, lastSentFilterArgs...))
		} else {
			ds = ds.Where(goqu.L(lastSentFilter))
		}
	}
	if len(eventFilters) > 0 {
		ds = ds.Where(goqu.L("event_filter = ANY(?)", pq.StringArray(eventFilters)))
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	log.Info(query)

	subMap := make(map[string][]types.Subscription, 0)
	err = db.FrontendWriterDB.Select(&subs, query, args...)
	if err != nil {
		return nil, err
	}

	log.Infof("Found %d subscriptions for event %s", len(subs), eventName)

	for _, sub := range subs {
		if _, ok := subMap[sub.EventFilter]; !ok {
			subMap[sub.EventFilter] = make([]types.Subscription, 0)
		}
		subMap[sub.EventFilter] = append(subMap[sub.EventFilter], sub)
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
