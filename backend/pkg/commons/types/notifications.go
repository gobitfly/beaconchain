package types

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type UserId uint64
type DashboardId uint64
type DashboardGroupId uint64
type ValidatorDashboardConfig struct {
	DashboardsById map[DashboardId]*ValidatorDashboard
}

type Subscription struct {
	ID          *uint64    `db:"id,omitempty"`
	UserID      *UserId    `db:"user_id,omitempty"`
	EventName   EventName  `db:"event_name"`
	EventFilter string     `db:"event_filter"`
	LastSent    *time.Time `db:"last_sent_ts"`
	LastEpoch   *uint64    `db:"last_sent_epoch"`
	// Channels        pq.StringArray `db:"channels"`
	CreatedTime    time.Time `db:"created_ts"`
	CreatedEpoch   uint64    `db:"created_epoch"`
	EventThreshold float64   `db:"event_threshold"`
	// State          sql.NullString `db:"internal_state" swaggertype:"string"`
	DashboardId        *int64 `db:"-"`
	DashboardName      string `db:"-"`
	DashboardGroupId   *int64 `db:"-"`
	DashboardGroupName string `db:"-"`
}

type ValidatorDashboard struct {
	Name   string `db:"name"`
	Groups map[DashboardGroupId]*ValidatorDashboardGroup
}

type ValidatorDashboardGroup struct {
	Name       string `db:"name"`
	Validators []types.ValidatorIndex
}
