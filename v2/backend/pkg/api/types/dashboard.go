package types

type AccountDashboard struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}
type ValidatorDashboard struct {
	Id             uint64        `json:"id"`
	Name           string        `json:"name"`
	Network        uint64        `json:"network"`
	PublicIds      []VDBPublicId `json:"public_ids,omitempty"`
	IsArchived     bool          `json:"is_archived"`
	ArchivedReason string        `json:"archived_reason,omitempty" tstype:"'user' | 'dashboard_limit' | 'validator_limit' | 'group_limit'"`
	ValidatorCount uint64        `json:"validator_count"`
	GroupCount     uint64        `json:"group_count"`
}

type UserDashboardsData struct {
	ValidatorDashboards []ValidatorDashboard `json:"validator_dashboards"`
	AccountDashboards   []AccountDashboard   `json:"account_dashboards"`
}

type GetUserDashboardsResponse ApiDataResponse[UserDashboardsData]
