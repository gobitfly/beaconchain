export type SummaryDetails = 'details_24h' | 'details_7d'| 'details_31d'| 'details_all'

export const SummaryDetailsEfficiencyProps = ['attestation_head', 'attestation_source', 'attestation_target', 'sync', 'proposals'] as const
export type SummaryDetailsEfficiencyProp = typeof SummaryDetailsEfficiencyProps[number]

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
  prev_cursor: string;
  next_cursor: string;
}
export interface VDBSummaryTableResponse {
  paging: Paging;
  data: VDBSummaryTableRow[];
}

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
