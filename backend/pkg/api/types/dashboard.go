package types

type AccountDashboard struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}
type ValidatorDashboard struct {
	Id        uint64        `json:"id"`
	Name      string        `json:"name"`
	PublicIds []VDBPublicId `json:"public_ids"`
}
type UserDashboardsData struct {
	ValidatorDashboards []ValidatorDashboard `json:"validator_dashboards"`
	AccountDashboards   []AccountDashboard   `json:"account_dashboards"`
}

type GetUserDashboardsResponse ApiDataResponse[UserDashboardsData]
