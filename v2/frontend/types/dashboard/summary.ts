import { type ChartHistorySeconds } from '~/types/api/common'

export const SummaryDetailsEfficiencyProps = [
  'attestations_head',
  'attestations_source',
  'attestations_target',
  'slashings',
  'sync',
] as const
export type SummaryDetailsEfficiencyProp =
  (typeof SummaryDetailsEfficiencyProps)[number]

export const SummaryDetailsEfficiencyValidatorProps = [
  'validators_sync',
  'validators_proposal',
  'validators_slashings',
] as const
export type SummaryDetailsEfficiencyValidatorProp =
  (typeof SummaryDetailsEfficiencyValidatorProps)[number]

export const SummaryDetailsEfficiencyLuckProps = [
  'proposal_luck',
  'sync_luck',
] as const
export type SummaryDetailsEfficiencyLuckProp =
  (typeof SummaryDetailsEfficiencyLuckProps)[number]

export const SummaryDetailsEfficiencyCustomProps = [ 'attestations' ] as const
export type SummaryDetailsEfficiencyCustomProp =
  (typeof SummaryDetailsEfficiencyCustomProps)[number]

export const SummaryDetailsEfficiencySpecialProps = [
  'reward',
  'efficiency',
  'apr',
  'luck',
  'attestation_avg_incl_dist',
  'attestation_efficiency',
  'proposals',
  'missed_rewards',
] as const
export type SummaryDetailsEfficiencySpecialProp =
  (typeof SummaryDetailsEfficiencySpecialProps)[number]

export type SummaryDetailsEfficiencyCombinedProp =
  | SummaryDetailsEfficiencyCustomProp
  | SummaryDetailsEfficiencyLuckProp
  | SummaryDetailsEfficiencyProp
  | SummaryDetailsEfficiencySpecialProp
  | SummaryDetailsEfficiencyValidatorProp

export type DashboardValidatorContext =
  | 'attestation'
  | 'dashboard'
  | 'group'
  | 'proposal'
  | 'slashings'
  | 'sync'

export type SummaryRow = {
  prop?: SummaryDetailsEfficiencyCombinedProp,
  title: string,
}

export const SummaryTimeFrames = [
  'last_1h',
  'last_24h',
  'last_7d',
  'last_30d',
  'all_time',
] as const
export type SummaryTimeFrame = (typeof SummaryTimeFrames)[number]

export type SummaryTableVisibility = {
  attestations: boolean,
  efficiency: boolean,
  proposals: boolean,
  reward: boolean,
  validatorsSortable: boolean,
}

export const SUMMARY_CHART_GROUP_TOTAL = -1
export const SUMMARY_CHART_GROUP_NETWORK_AVERAGE = -2

export type AggregationTimeframe = keyof ChartHistorySeconds
export const AggregationTimeframes: AggregationTimeframe[] = [
  'epoch',
  'hourly',
  'daily',
  'weekly',
]

export const EfficiencyTypes = [
  'all',
  'attestation',
  'sync',
  'proposal',
]
export type EfficiencyType = (typeof EfficiencyTypes)[number]

export type SummaryChartFilter = {
  aggregation: AggregationTimeframe,
  efficiency: EfficiencyType,
  groupIds: number[],
  initialised?: boolean,
}
