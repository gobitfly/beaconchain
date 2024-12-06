package types

import "github.com/gobitfly/beaconchain/pkg/api/enums"

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
