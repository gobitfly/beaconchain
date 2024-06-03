// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { PeriodicValues, ClElValue, ApiDataResponse, ApiPagingResponse, StatusCount, Luck, ChartData, ValidatorHistoryDuties, Address, PubKey, Hash } from './common'

//////////
// source: validator_dashboard.go

/**
 * ------------------------------------------------------------
 * Overview
 */
export interface VDBOverviewValidators {
  online: number /* uint64 */;
  offline: number /* uint64 */;
  pending: number /* uint64 */;
  exited: number /* uint64 */;
  slashed: number /* uint64 */;
}
export interface VDBOverviewGroup {
  id: number /* uint64 */;
  name?: string;
  count: number /* uint64 */;
}
export interface VDBOverviewData {
  name: string;
  groups: VDBOverviewGroup[];
  validators: VDBOverviewValidators;
  efficiency: PeriodicValues<number /* float64 */>;
  rewards: PeriodicValues<ClElValue<string /* decimal.Decimal */>>;
  apr: PeriodicValues<ClElValue<number /* float64 */>>;
}
export type InternalGetValidatorDashboardResponse = ApiDataResponse<VDBOverviewData>;
/**
 * ------------------------------------------------------------
 * Summary Tab
 */
export interface VDBSummaryTableRow {
  group_id: number /* uint64 */;
  efficiency: PeriodicValues<number /* float64 */>;
  validators: number /* uint64 */[];
}
export type InternalGetValidatorDashboardSummaryResponse = ApiPagingResponse<VDBSummaryTableRow>;
export interface VDBGroupSummaryColumnItem {
  status_count: StatusCount;
  validators?: number /* uint64 */[];
}
export interface VDBGroupSummaryColumn {
  attestations_head: VDBGroupSummaryColumnItem;
  attestations_source: VDBGroupSummaryColumnItem;
  attestations_target: VDBGroupSummaryColumnItem;
  attestation_count: StatusCount;
  attestation_efficiency: number /* float64 */;
  attestation_avg_incl_dist: number /* float64 */;
  sync: VDBGroupSummaryColumnItem;
  proposals: VDBGroupSummaryColumnItem;
  slashed: VDBGroupSummaryColumnItem; // Failed slashings are count of validators in the group that were slashed
  apr: ClElValue<number /* float64 */>;
  income: ClElValue<string /* decimal.Decimal */>;
  luck: Luck;
}
export interface VDBGroupSummaryData {
  last_24h: VDBGroupSummaryColumn;
  last_7d: VDBGroupSummaryColumn;
  last_30d: VDBGroupSummaryColumn;
  all_time: VDBGroupSummaryColumn;
}
export type InternalGetValidatorDashboardGroupSummaryResponse = ApiDataResponse<VDBGroupSummaryData>;
export type InternalGetValidatorDashboardSummaryChartResponse = ApiDataResponse<ChartData<number /* int */, number /* float64 */>>; // line chart, series id is group id
export type InternalGetValidatorDashboardValidatorIndicesResponse = ApiDataResponse<number /* uint64 */[]>;
/**
 * ------------------------------------------------------------
 * Rewards Tab
 */
export interface VDBRewardesTableDuty {
  attestation?: number /* float64 */;
  proposal?: number /* float64 */;
  sync?: number /* float64 */;
  slashing?: number /* uint64 */;
}
export interface VDBRewardsTableRow {
  epoch: number /* uint64 */;
  duty: VDBRewardesTableDuty;
  group_id: number /* int64 */;
  reward: ClElValue<string /* decimal.Decimal */>;
}
export type InternalGetValidatorDashboardRewardsResponse = ApiPagingResponse<VDBRewardsTableRow>;
export interface VDBGroupRewardsDetails {
  status_count: StatusCount;
  income: string /* decimal.Decimal */;
}
export interface VDBGroupRewardsData {
  attestations_source: VDBGroupRewardsDetails;
  attestations_target: VDBGroupRewardsDetails;
  attestations_head: VDBGroupRewardsDetails;
  sync: VDBGroupRewardsDetails;
  slashing: VDBGroupRewardsDetails;
  inactivity: VDBGroupRewardsDetails;
  proposal: VDBGroupRewardsDetails;
  proposal_el_reward: string /* decimal.Decimal */;
  proposal_cl_att_inc_reward: string /* decimal.Decimal */;
  proposal_cl_sync_inc_reward: string /* decimal.Decimal */;
  proposal_cl_slashing_inc_reward: string /* decimal.Decimal */;
}
export type InternalGetValidatorDashboardGroupRewardsResponse = ApiDataResponse<VDBGroupRewardsData>;
export type InternalGetValidatorDashboardRewardsChartResponse = ApiDataResponse<ChartData<number /* int */, string /* decimal.Decimal */>>; // bar chart, series id is group id, property is 'el' or 'cl'
export interface VDBEpochDutiesTableRow {
  validator: number /* uint64 */;
  duties: ValidatorHistoryDuties;
}
export type InternalGetValidatorDashboardDutiesResponse = ApiPagingResponse<VDBEpochDutiesTableRow>;
/**
 * ------------------------------------------------------------
 * Blocks Tab
 */
export interface VDBBlocksTableRow {
  proposer: number /* uint64 */;
  group_id: number /* uint64 */;
  epoch: number /* uint64 */;
  slot: number /* uint64 */;
  block: number /* uint64 */;
  status: 'success' | 'missed' | 'orphaned' | 'scheduled';
  reward_recipient: Address;
  reward: ClElValue<string /* decimal.Decimal */>;
  graffiti: string;
}
export type InternalGetValidatorDashboardBlocksResponse = ApiPagingResponse<VDBBlocksTableRow>;
export interface VDBHeatmapEvents {
  proposal: boolean;
  slash: boolean;
  sync: boolean;
}
export interface VDBHeatmapCell {
  x: number /* int64 */; // Timestamp
  y: number /* uint64 */; // Group ID
  value: number /* float64 */; // Attestaton Rewards
  events?: VDBHeatmapEvents;
}
export interface VDBHeatmap {
  timestamps: number /* int64 */[]; // X-Axis Categories (unix timestamp)
  group_ids: number /* uint64 */[]; // Y-Axis Categories
  data: VDBHeatmapCell[];
  aggregation: 'epoch' | 'day';
}
export type InternalGetValidatorDashboardHeatmapResponse = ApiDataResponse<VDBHeatmap>;
export interface VDBHeatmapTooltipData {
  timestamp: number /* int64 */; // epoch or day
  proposers: StatusCount;
  syncs: number /* uint64 */;
  slashings: StatusCount;
  attestations_head: StatusCount;
  attestations_source: StatusCount;
  attestations_target: StatusCount;
  attestation_income: string /* decimal.Decimal */;
  attestation_efficiency: number /* float64 */;
}
export type InternalGetValidatorDashboardGroupHeatmapResponse = ApiDataResponse<VDBHeatmapTooltipData>;
/**
 * ------------------------------------------------------------
 * Deposits Tab
 */
export interface VDBExecutionDepositsTableRow {
  public_key: PubKey;
  index?: number /* uint64 */;
  group_id: number /* uint64 */;
  block: number /* uint64 */;
  timestamp: number /* int64 */;
  from: Address;
  depositor: Address;
  tx_hash: Hash;
  withdrawal_credential: Hash;
  amount: string /* decimal.Decimal */;
  valid: boolean;
}
export type InternalGetValidatorDashboardExecutionLayerDepositsResponse = ApiPagingResponse<VDBExecutionDepositsTableRow>;
export interface VDBConsensusDepositsTableRow {
  public_key: PubKey;
  index: number /* uint64 */;
  group_id: number /* uint64 */;
  epoch: number /* uint64 */;
  slot: number /* uint64 */;
  withdrawal_credential: Hash;
  amount: string /* decimal.Decimal */;
  signature: Hash;
}
export type InternalGetValidatorDashboardConsensusLayerDepositsResponse = ApiPagingResponse<VDBConsensusDepositsTableRow>;
export interface VDBTotalExecutionDepositsData {
  total_amount: string /* decimal.Decimal */;
}
export type InternalGetValidatorDashboardTotalExecutionDepositsResponse = ApiDataResponse<VDBTotalExecutionDepositsData>;
export interface VDBTotalConsensusDepositsData {
  total_amount: string /* decimal.Decimal */;
}
export type InternalGetValidatorDashboardTotalConsensusDepositsResponse = ApiDataResponse<VDBTotalConsensusDepositsData>;
/**
 * ------------------------------------------------------------
 * Withdrawals Tab
 */
export interface VDBWithdrawalsTableRow {
  epoch: number /* uint64 */;
  slot: number /* uint64 */;
  index: number /* uint64 */;
  group_id: number /* uint64 */;
  recipient: Address;
  amount: string /* decimal.Decimal */;
  is_missing_estimate: boolean;
}
export type InternalGetValidatorDashboardWithdrawalsResponse = ApiPagingResponse<VDBWithdrawalsTableRow>;
export interface VDBTotalWithdrawalsData {
  total_amount: string /* decimal.Decimal */;
}
export type InternalGetValidatorDashboardTotalWithdrawalsResponse = ApiDataResponse<VDBTotalWithdrawalsData>;
/**
 * ------------------------------------------------------------
 * Manage Modal
 */
export interface VDBManageValidatorsTableRow {
  index: number /* uint64 */;
  public_key: PubKey;
  group_id: number /* uint64 */;
  balance: string /* decimal.Decimal */;
  status: 'pending' | 'online' | 'offline' | 'exiting' | 'exited' | 'slashed' | 'withdrawn';
  queue_position?: number /* uint64 */;
  withdrawal_credential: Hash;
}
export type InternalGetValidatorDashboardValidatorsResponse = ApiPagingResponse<VDBManageValidatorsTableRow>;
/**
 * ------------------------------------------------------------
 * Misc.
 */
export interface VDBPostReturnData {
  id: number /* uint64 */;
  user_id: number /* uint64 */;
  name: string;
  network: number /* uint64 */;
  created_at: number /* int64 */;
}
export interface VDBPostCreateGroupData {
  id: number /* uint64 */;
  name: string;
}
export interface VDBPostValidatorsData {
  public_key: string;
  group_id: number /* uint64 */;
}
