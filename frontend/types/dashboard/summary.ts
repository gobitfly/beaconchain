export const SummaryDetailsEfficiencyProps = ['attestations_head', 'attestations_source', 'attestations_target', 'slashings'] as const
export type SummaryDetailsEfficiencyProp = typeof SummaryDetailsEfficiencyProps[number]

export const SummaryDetailsEfficiencyValidatorProps = ['validators_sync', 'validators_proposal', 'validators_slashings'] as const
export type SummaryDetailsEfficiencyValidatorProp = typeof SummaryDetailsEfficiencyValidatorProps[number]

export const SummaryDetailsEfficiencyLuckProps = ['proposal_luck', 'sync_luck'] as const
export type SummaryDetailsEfficiencyLuckProp = typeof SummaryDetailsEfficiencyLuckProps[number]

export const SummaryDetailsEfficiencyCustomProps = ['attestation_total'] as const
export type SummaryDetailsEfficiencyCustomProp = typeof SummaryDetailsEfficiencyCustomProps[number]

export const SummaryDetailsEfficiencySpecialProps = ['reward', 'sync', 'efficiency', 'apr', 'luck', 'attestation_avg_incl_dist', 'attestation_efficiency', 'proposals'] as const
export type SummaryDetailsEfficiencySpecialProp = typeof SummaryDetailsEfficiencySpecialProps[number]

export type SummaryDetailsEfficiencyCombinedProp = SummaryDetailsEfficiencySpecialProp | SummaryDetailsEfficiencyProp | SummaryDetailsEfficiencyCustomProp | SummaryDetailsEfficiencyLuckProp | SummaryDetailsEfficiencyValidatorProp

export type DashboardValidatorContext = 'dashboard' | 'group' | 'attestation' | 'sync' | 'slashings' | 'proposal'

export type SummaryRow = { prop: SummaryDetailsEfficiencyCombinedProp, title: string}

export const SummaryTimeFrames = ['last_1h', 'last_24h', 'last_7d', 'last_30d', 'all_time'] as const
export type SummaryTimeFrame = typeof SummaryTimeFrames[number]
