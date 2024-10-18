package notification

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

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
func GetSubsForEventFilter(eventName types.EventName, lastSentFilter string, lastSentFilterArgs []interface{}, eventFilters []string, validatorDashboardConfig *types.ValidatorDashboardConfig) (map[string][]types.Subscription, error) {
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

	eventNameForQuery := utils.GetNetwork() + ":" + string(eventName)

	if _, ok := types.UserIndexEventsMap[eventName]; ok {
		eventNameForQuery = string(eventName)
	}
	ds := goqu.Dialect("postgres").From("users_subscriptions").Select(
		goqu.T("users_subscriptions").Col("id"),
		goqu.C("user_id"),
		goqu.C("event_filter"),
		goqu.C("last_sent_epoch"),
		goqu.C("created_epoch"),
		goqu.C("event_threshold"),
		goqu.C("event_name"),
	).Join(goqu.T("users"), goqu.On(goqu.T("users").Col("id").Eq(goqu.T("users_subscriptions").Col("user_id")))).
		Where(goqu.L("(event_name = ? AND user_id <> 0)", eventNameForQuery)).
		Where(goqu.L("(users.notifications_do_not_disturb_ts IS NULL OR users.notifications_do_not_disturb_ts < NOW())")).
		// filter out users that have all notification channels disabled
		Where(goqu.L("(select bool_or(active) from users_notification_channels where users_notification_channels.user_id = users_subscriptions.user_id)"))

	if lastSentFilter != "" {
		if len(lastSentFilterArgs) > 0 {
			ds = ds.Where(goqu.L(lastSentFilter, lastSentFilterArgs...))
		} else {
			ds = ds.Where(goqu.L(lastSentFilter))
		}
	}
	if len(eventFilters) > 0 {
		ds = ds.Where(goqu.L("(event_filter = ANY(?))", pq.StringArray(eventFilters)))
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	subMap := make(map[string][]types.Subscription, 0)
	err = db.FrontendWriterDB.Select(&subs, query, args...)
	if err != nil {
		return nil, err
	}

	log.Infof("found %d subscriptions for event %s", len(subs), eventName)
	// zero := uint64(0)
	for _, sub := range subs {
		// sub.LastEpoch = &zero
		// sub.LastSent = &time.Time{}
		sub.EventName = types.EventName(strings.Replace(string(sub.EventName), utils.GetNetwork()+":", "", 1)) // remove the network name from the event name
		if strings.HasPrefix(sub.EventFilter, "vdb:") {
			dashboardData := strings.Split(sub.EventFilter, ":")
			if len(dashboardData) != 3 {
				log.Error(fmt.Errorf("invalid dashboard subscription: %s", sub.EventFilter), "invalid dashboard subscription", 0)
				continue
			}
			dashboardId, err := strconv.ParseInt(dashboardData[1], 10, 64)
			if err != nil {
				log.Error(err, "Invalid dashboard subscription", 0)
				continue
			}
			sub.DashboardId = &dashboardId

			dashboardGroupId, err := strconv.ParseInt(dashboardData[2], 10, 64)
			if err != nil {
				log.Error(err, "Invalid dashboard subscription", 0)
				continue
			}
			sub.DashboardGroupId = &dashboardGroupId
			if dashboard, ok := validatorDashboardConfig.DashboardsById[types.DashboardId(dashboardId)]; ok {
				if dashboard.Name == "" {
					dashboard.Name = fmt.Sprintf("Dashboard %d", dashboardId)
				}
				if group, ok := dashboard.Groups[types.DashboardGroupId(dashboardGroupId)]; ok {
					if group.Name == "" {
						group.Name = "default"
					}

					uniqueRPLNodes := make(map[string]struct{})

					for _, validatorIndex := range group.Validators {
						validatorEventFilterRaw, err := GetPubkeyForIndex(validatorIndex)
						if err != nil {
							log.Error(err, "error retrieving pubkey for validator", 0, map[string]interface{}{"validator": validatorIndex})
							continue
						}
						validatorEventFilter := hex.EncodeToString(validatorEventFilterRaw)

						if eventName == types.RocketpoolCollateralMaxReachedEventName || eventName == types.RocketpoolCollateralMinReachedEventName {
							// Those two RPL notifications are not tied to a specific validator but to a node address, create a subscription for each
							// node in the group
							nodeAddress, ok := validatorDashboardConfig.RocketpoolNodeByPubkey[validatorEventFilter]
							if !ok {
								// Validator is not a rocketpool minipool
								continue
							}
							if _, ok := uniqueRPLNodes[nodeAddress]; !ok {
								if _, ok := subMap[nodeAddress]; !ok {
									subMap[nodeAddress] = make([]types.Subscription, 0)
								}
								hydratedSub := types.Subscription{
									ID:                 sub.ID,
									UserID:             sub.UserID,
									EventName:          sub.EventName,
									EventFilter:        nodeAddress,
									LastSent:           sub.LastSent,
									LastEpoch:          sub.LastEpoch,
									CreatedTime:        sub.CreatedTime,
									CreatedEpoch:       sub.CreatedEpoch,
									EventThreshold:     sub.EventThreshold,
									DashboardId:        sub.DashboardId,
									DashboardName:      dashboard.Name,
									DashboardGroupId:   sub.DashboardGroupId,
									DashboardGroupName: group.Name,
								}
								subMap[nodeAddress] = append(subMap[nodeAddress], hydratedSub)
							}
							uniqueRPLNodes[nodeAddress] = struct{}{}
						} else {
							if _, ok := subMap[validatorEventFilter]; !ok {
								subMap[validatorEventFilter] = make([]types.Subscription, 0)
							}
							hydratedSub := types.Subscription{
								ID:                 sub.ID,
								UserID:             sub.UserID,
								EventName:          sub.EventName,
								EventFilter:        validatorEventFilter,
								LastSent:           sub.LastSent,
								LastEpoch:          sub.LastEpoch,
								CreatedTime:        sub.CreatedTime,
								CreatedEpoch:       sub.CreatedEpoch,
								EventThreshold:     sub.EventThreshold,
								DashboardId:        sub.DashboardId,
								DashboardName:      dashboard.Name,
								DashboardGroupId:   sub.DashboardGroupId,
								DashboardGroupName: group.Name,
							}
							subMap[validatorEventFilter] = append(subMap[validatorEventFilter], hydratedSub)
						}

						//log.Infof("hydrated subscription for validator %v of dashboard %d and group %d for user %d", hydratedSub.EventFilter, *hydratedSub.DashboardId, *hydratedSub.DashboardGroupId, *hydratedSub.UserID)
					}
				}
			}
		} else {
			if _, ok := subMap[sub.EventFilter]; !ok {
				subMap[sub.EventFilter] = make([]types.Subscription, 0)
			}
			subMap[sub.EventFilter] = append(subMap[sub.EventFilter], sub)
		}
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
