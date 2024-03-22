// can be ether a dashboard id or a list of validators
export type DashboardKey = number | string

export type DashboardType = 'validator' | 'account'

// TODO: once the search PR is finished check if we can get these from somewhere else
export type ValidatorDashboardNetwork = 'ethereum' | 'gnosis'

export const DAHSHBOARDS_ALL_GROUPS_ID = -1
