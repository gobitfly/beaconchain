package types

// count indicator per block details tab; each tab is only present if count > 0
type ArchiverDashboard struct {
	DashboardId    uint64
	IsArchived     bool
	GroupCount     uint64
	ValidatorCount uint64
}
