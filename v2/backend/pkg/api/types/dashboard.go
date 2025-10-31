package types

type AccountDashboard struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}
type ValidatorDashboard struct {
	Id             uint64        `json:"id" extensions:"x-order=1"`
	Name           string        `json:"name" extensions:"x-order=2"`
	Network        uint64        `json:"network" extensions:"x-order=3"`
	PublicIds      []VDBPublicId `json:"public_ids,omitempty" extensions:"x-order=4"`
	IsArchived     bool          `json:"is_archived" extensions:"x-order=5"`
	ArchivedReason string        `json:"archived_reason,omitempty" tstype:"'user' | 'dashboard_limit' | 'validator_limit' | 'group_limit'" extensions:"x-order=6"`
	ValidatorCount uint64        `json:"validator_count" extensions:"x-order=7"`
	GroupCount     uint64        `json:"group_count" extensions:"x-order=8"`
}

type UserDashboardsData struct {
	ValidatorDashboards []ValidatorDashboard `json:"validator_dashboards"`
	AccountDashboards   []AccountDashboard   `json:"account_dashboards"`
}

type GetUserDashboardsResponse ApiDataResponse[UserDashboardsData]
