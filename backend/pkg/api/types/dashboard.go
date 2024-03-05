package types

type Dashboard struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type UserDashboardsData struct {
	ValidatorDashboards []Dashboard `json:"validator_dashboards"`
	AccountDashboards   []Dashboard `json:"account_dashboards"`
}

type GetUserDashboardsResponse ApiDataResponse[UserDashboardsData]
