package types

import (
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/shopspring/decimal"
)

type ArchiverDashboard struct {
	DashboardId    uint64
	IsArchived     bool
	GroupCount     uint64
	ValidatorCount uint64
}

type ArchiverDashboardArchiveReason struct {
	DashboardId    uint64
	ArchivedReason enums.VDBArchivedReason
}

// TODO: Find a good place for this
type RpMinipoolInfo struct {
	NodeFee             float64
	NodeDepositBalance  decimal.Decimal
	UserDepositBalance  decimal.Decimal
	SmoothingPoolReward map[uint64]decimal.Decimal
}
