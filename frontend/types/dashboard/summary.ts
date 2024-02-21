export const SummaryDetails = ['details_24h', 'details_7d', 'details_31d', 'details_all'] as const
export type SummaryDetail = typeof SummaryDetails[number]

export const SummaryDetailsEfficiencyProps = ['attestation_head', 'attestation_source', 'attestation_target', 'sync', 'proposals', 'slashings'] as const
export type SummaryDetailsEfficiencyProp = typeof SummaryDetailsEfficiencyProps[number]

export const SummaryDetailsEfficiencyValidatorProps = ['validators_sync', 'validators_proposal', 'validators_slashings', 'validators_attestation'] as const
export type SummaryDetailsEfficiencyValidatorProp = typeof SummaryDetailsEfficiencyValidatorProps[number]

export const SummaryDetailsEfficiencyLuckProps = ['proposal_luck', 'sync_luck'] as const
export type SummaryDetailsEfficiencyLuckProp = typeof SummaryDetailsEfficiencyLuckProps[number]

export const SummaryDetailsEfficiencyCustomProps = ['attestation_total'] as const
export type SummaryDetailsEfficiencyCustomProp = typeof SummaryDetailsEfficiencyCustomProps[number]

export const SummaryDetailsEfficiencySpecialProps = ['efficiency_total', 'apr', 'luck', 'attestation_avg_incl_dist', 'attestation_efficiency'] as const
export type SummaryDetailsEfficiencySpecialProp = typeof SummaryDetailsEfficiencySpecialProps[number]

export type SummaryDetailsEfficiencyCombinedProp = SummaryDetailsEfficiencySpecialProp | SummaryDetailsEfficiencyProp | SummaryDetailsEfficiencyCustomProp | SummaryDetailsEfficiencyLuckProp | SummaryDetailsEfficiencyValidatorProp

export type DashboardValidatorContext = 'dashboard' | 'group' | 'attestation' | 'sync' | 'slashings' | 'proposal'

export type SummaryRow = { details: SummaryDetail[], prop: SummaryDetailsEfficiencyCombinedProp, title: string, className: string }

// TODO: Replace with types below with the ones we get from the backend

export interface VDBSummaryTableRow {
  group_id: number;
  efficiency_24h: number;
  efficiency_7d: number;
  efficiency_31d: number;
  efficiency_all: number;
  validators: number[];
}
export interface Paging {
  prev_cursor?: string;
  next_cursor?: string;
  last_cursor?: string;
  total_count?: number;
}

export interface TableResponse<T> {
  paging: Paging;
  data: T[];
}

export interface VDBSummaryTableResponse extends TableResponse<VDBSummaryTableRow>{}

export interface Luck {
  percent: number;
  expected: number;
  average: number;
}
export interface ClEl {
  el: string;
  cl: string;
}
export interface ClElFloat {
  el: number;
  cl: number;
}
export interface Count {
  success: number;
  failed: number;
}
export interface VDBGroupSummaryColumnItem {
  count: Count;
  earned: string;
  penalty: string;
  validators?: number[];
}
export interface VDBGroupSummaryColumn {
  attestation_head: VDBGroupSummaryColumnItem;
  attestation_source: VDBGroupSummaryColumnItem;
  attestation_target: VDBGroupSummaryColumnItem;
  attestation_efficiency: number;
  attestation_avg_incl_dist: number;
  sync: VDBGroupSummaryColumnItem;
  proposals: VDBGroupSummaryColumnItem;
  slashings: VDBGroupSummaryColumnItem;
  apr: ClElFloat;
  income: ClEl;
  proposal_luck: Luck;
  sync_luck: Luck;
}
export interface VDBGroupSummaryResponse {
  details_24h: VDBGroupSummaryColumn;
  details_7d: VDBGroupSummaryColumn;
  details_31d: VDBGroupSummaryColumn;
  details_all: VDBGroupSummaryColumn;
}
