/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface HandlersPublicPostUserNotificationsTestWebhookRequest {
  is_webhook_discord_enabled?: boolean;
  webhook_url?: string;
}

export interface HandlersPublicPostValidatorDashboardPublicIdsRequest {
  name?: string;
  share_settings?: {
    share_groups?: boolean;
  };
}

export interface HandlersPublicPostValidatorDashboardValidatorBulkDeletionsRequest {
  validators?: HandlersIntOrString[];
}

export interface HandlersPublicPostValidatorDashboardValidatorsRequest {
  deposit_address?: string;
  graffiti?: string;
  group_id?: number;
  validators?: HandlersIntOrString[];
  withdrawal_credential?: string;
}

export interface HandlersPublicPostValidatorDashboardsRequest {
  name?: string;
  network?: "ethereum" | "gnosis";
}

export interface HandlersPublicPutUserNotificationSettingsAccountDashboardRequest {
  /** 0 does not disable, is_erc20_token_transfers_subscribed determines if it's enabled */
  erc20_token_transfers_value_threshold?: number;
  is_erc1155_token_transfers_subscribed?: boolean;
  is_erc20_token_transfers_subscribed?: boolean;
  is_erc721_token_transfers_subscribed?: boolean;
  is_ignore_spam_transactions_enabled?: boolean;
  is_incoming_transactions_subscribed?: boolean;
  is_outgoing_transactions_subscribed?: boolean;
  is_webhook_discord_enabled?: boolean;
  subscribed_chain_ids?: HandlersIntOrString[];
  webhook_url?: string;
}

export interface HandlersPublicPutUserNotificationSettingsClientRequest {
  is_subscribed?: boolean;
}

export interface HandlersPublicPutUserNotificationSettingsNetworksRequest {
  gas_above_threshold?: string;
  gas_below_threshold?: string;
  is_gas_above_subscribed?: boolean;
  is_gas_below_subscribed?: boolean;
  is_new_reward_round_subscribed?: boolean;
  is_participation_rate_subscribed?: boolean;
  participation_rate_threshold?: number;
}

export interface HandlersPublicPutUserNotificationSettingsPairedDevicesRequest {
  is_notifications_enabled?: boolean;
  name?: string;
}

export interface HandlersPublicPutValidatorDashboardArchivingRequest {
  is_archived?: boolean;
}

export interface HandlersPublicPutValidatorDashboardGroupsRequest {
  name?: string;
}

export interface HandlersPublicPutValidatorDashboardNameRequest {
  name?: string;
}

export interface HandlersPublicPutValidatorDashboardPublicIdRequest {
  name?: string;
  share_settings?: {
    share_groups?: boolean;
  };
}

export type HandlersIntOrString = object;

export interface TypesAccountDashboard {
  id?: number;
  name?: string;
}

export interface TypesAddress {
  ens?: string;
  hash?: string;
  is_contract?: boolean;
  label?: string;
}

export interface TypesApiDataResponseArrayTypesVDBPostValidatorsData {
  data?: TypesVDBPostValidatorsData[];
}

export interface TypesApiDataResponseTypesUserDashboardsData {
  data?: TypesUserDashboardsData;
}

export interface TypesApiDataResponseTypesVDBPostArchivingReturnData {
  data?: TypesVDBPostArchivingReturnData;
}

export interface TypesApiDataResponseTypesVDBPostCreateGroupData {
  data?: TypesVDBPostCreateGroupData;
}

export interface TypesApiDataResponseTypesVDBPostReturnData {
  data?: TypesVDBPostReturnData;
}

export interface TypesApiDataResponseTypesVDBPublicId {
  data?: TypesVDBPublicId;
}

export interface TypesApiErrorResponse {
  error?: string;
}

export interface TypesChartDataIntDecimalDecimal {
  /** x-axis */
  categories?: number[];
  series?: TypesChartSeriesIntDecimalDecimal[];
}

export interface TypesChartDataIntFloat64 {
  /** x-axis */
  categories?: number[];
  series?: TypesChartSeriesIntFloat64[];
}

export interface TypesChartHistorySeconds {
  daily?: number;
  epoch?: number;
  hourly?: number;
  weekly?: number;
}

export interface TypesChartSeriesIntDecimalDecimal {
  /** y-axis values */
  data?: number[];
  /** id may be a string or an int */
  id?: number;
  /** for stacking bar charts */
  property?: string;
}

export interface TypesChartSeriesIntFloat64 {
  /** y-axis values */
  data?: number[];
  /** id may be a string or an int */
  id?: number;
  /** for stacking bar charts */
  property?: string;
}

export interface TypesClElValueDecimalDecimal {
  cl?: number;
  el?: number;
}

export interface TypesClElValueFloat64 {
  cl?: number;
  el?: number;
}

export interface TypesGetUserNotificationClientsResponse {
  data?: TypesNotificationClientsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetUserNotificationDashboardsResponse {
  data?: TypesNotificationDashboardsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetUserNotificationMachinesResponse {
  data?: TypesNotificationMachinesTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetUserNotificationNetworksResponse {
  data?: TypesNotificationNetworksTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetUserNotificationSettingsDashboardsResponse {
  data?: TypesNotificationSettingsDashboardsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetUserNotificationSettingsResponse {
  data?: TypesNotificationSettings;
}

export interface TypesGetUserNotificationsAccountDashboardResponse {
  data?: TypesNotificationAccountDashboardDetail;
}

export interface TypesGetUserNotificationsResponse {
  data?: TypesNotificationOverviewData;
}

export interface TypesGetUserNotificationsValidatorDashboardResponse {
  data?: TypesNotificationValidatorDashboardDetail;
}

export interface TypesGetValidatorDashboardBlocksResponse {
  data?: TypesVDBBlocksTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardConsensusLayerDepositsResponse {
  data?: TypesVDBConsensusDepositsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardDutiesResponse {
  data?: TypesVDBEpochDutiesTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardExecutionLayerDepositsResponse {
  data?: TypesVDBExecutionDepositsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardGroupRewardsResponse {
  data?: TypesVDBGroupRewardsData;
}

export interface TypesGetValidatorDashboardGroupSummaryResponse {
  data?: TypesVDBGroupSummaryData;
}

export interface TypesGetValidatorDashboardResponse {
  data?: TypesVDBOverviewData;
}

export interface TypesGetValidatorDashboardRewardsChartResponse {
  data?: TypesChartDataIntDecimalDecimal;
}

export interface TypesGetValidatorDashboardRewardsResponse {
  data?: TypesVDBRewardsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardRocketPoolMinipoolsResponse {
  data?: TypesVDBRocketPoolMinipoolsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardRocketPoolResponse {
  data?: TypesVDBRocketPoolTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardSlotVizResponse {
  data?: TypesSlotVizEpoch[];
}

export interface TypesGetValidatorDashboardSummaryChartResponse {
  data?: TypesChartDataIntFloat64;
}

export interface TypesGetValidatorDashboardSummaryResponse {
  data?: TypesVDBSummaryTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardSummaryValidatorsResponse {
  data?: TypesVDBSummaryValidatorsData[];
}

export interface TypesGetValidatorDashboardTotalConsensusDepositsResponse {
  data?: TypesVDBTotalConsensusDepositsData;
}

export interface TypesGetValidatorDashboardTotalExecutionDepositsResponse {
  data?: TypesVDBTotalExecutionDepositsData;
}

export interface TypesGetValidatorDashboardTotalRocketPoolResponse {
  data?: TypesVDBRocketPoolTableRow;
}

export interface TypesGetValidatorDashboardTotalWithdrawalsResponse {
  data?: TypesVDBTotalWithdrawalsData;
}

export interface TypesGetValidatorDashboardValidatorsResponse {
  data?: TypesVDBManageValidatorsTableRow[];
  paging?: TypesPaging;
}

export interface TypesGetValidatorDashboardWithdrawalsResponse {
  data?: TypesVDBWithdrawalsTableRow[];
  paging?: TypesPaging;
}

export interface TypesIndexBlocks {
  blocks?: number[];
  index?: number;
}

export interface TypesIndexEpoch {
  epoch?: number;
  index?: number;
}

export interface TypesIndexSlots {
  index?: number;
  slots?: number[];
}

export interface TypesLuck {
  proposal?: TypesLuckItem;
  sync?: TypesLuckItem;
}

export interface TypesLuckItem {
  average_interval_seconds?: number;
  expected_timestamp?: number;
  percent?: number;
}

export interface TypesNotificationAccountDashboardDetail {
  erc1155_token_transfers?: TypesNotificationEventExecution[];
  erc20_token_transfers?: TypesNotificationEventExecution[];
  erc721_token_transfers?: TypesNotificationEventExecution[];
  incoming_transactions?: TypesNotificationEventExecution[];
  outgoing_transactions?: TypesNotificationEventExecution[];
}

export interface TypesNotificationClientsTableRow {
  client_name?: string;
  timestamp?: number;
  url?: string;
  version?: string;
}

export interface TypesNotificationDashboardsTableRow {
  chain_id?: number;
  dashboard_id?: number;
  dashboard_name?: string;
  entity_count?: number;
  epoch?: number;
  event_types?: string[];
  group_id?: number;
  group_name?: string;
  /** if false it's a validator dashboard */
  is_account_dashboard?: boolean;
}

export interface TypesNotificationEventExecution {
  address?: TypesAddress;
  amount?: number;
  /** this field will prob change depending on how execution stuff is implemented */
  token_name?: string;
  transaction_hash?: string;
}

export interface TypesNotificationEventValidatorBackOnline {
  epoch_count?: number;
  index?: number;
}

export interface TypesNotificationEventWithdrawal {
  address?: TypesAddress;
  amount?: number;
  index?: number;
}

export interface TypesNotificationMachinesTableRow {
  event_type?: string;
  machine_name?: string;
  threshold?: number;
  timestamp?: number;
}

export interface TypesNotificationNetwork {
  chain_id?: number;
  settings?: TypesNotificationSettingsNetwork;
}

export interface TypesNotificationNetworksTableRow {
  chain_id?: number;
  event_type?: string;
  /** participation rate threshold should also be passed as decimal string */
  threshold?: number;
  timestamp?: number;
}

export interface TypesNotificationOverviewData {
  adb_most_notified_groups?: string[];
  adb_subscriptions_count?: number;
  clients_subscription_count?: number;
  is_email_notifications_enabled?: boolean;
  is_push_notifications_enabled?: boolean;
  /** daily limit should be available in user info */
  last_24h_email_count?: number;
  last_24h_push_count?: number;
  last_24h_webhook_count?: number;
  machines_subscription_count?: number;
  networks_subscription_count?: number;
  next_email_count_reset_timestamp?: number;
  /** these will list 3 group names */
  vdb_most_notified_groups?: string[];
  /** counts are shown in their respective tables */
  vdb_subscriptions_count?: number;
}

export interface TypesNotificationPairedDevice {
  id?: number;
  is_notifications_enabled?: boolean;
  name?: string;
  paired_timestamp?: number;
}

export interface TypesNotificationSettings {
  clients?: TypesNotificationSettingsClient[];
  general_settings?: TypesNotificationSettingsGeneral;
  has_machines?: boolean;
  networks?: TypesNotificationNetwork[];
  paired_devices?: TypesNotificationPairedDevice[];
}

export interface TypesNotificationSettingsAccountDashboard {
  erc20_token_transfers_value_threshold?: number;
  is_erc1155_token_transfers_subscribed?: boolean;
  is_erc20_token_transfers_subscribed?: boolean;
  is_erc721_token_transfers_subscribed?: boolean;
  is_ignore_spam_transactions_enabled?: boolean;
  is_incoming_transactions_subscribed?: boolean;
  is_outgoing_transactions_subscribed?: boolean;
  is_webhook_discord_enabled?: boolean;
  subscribed_chain_ids?: number[];
  webhook_url?: string;
}

export interface TypesNotificationSettingsClient {
  category?: string;
  id?: number;
  is_subscribed?: boolean;
  name?: string;
}

export interface TypesNotificationSettingsDashboardsTableRow {
  chain_ids?: number[];
  dashboard_id?: number;
  dashboard_name?: string;
  group_id?: number;
  group_name?: string;
  /** if false it's a validator dashboard */
  is_account_dashboard?: boolean;
  is_archived?: boolean;
  /** if it's a validator dashboard, Settings is NotificationSettingsAccountDashboard, otherwise NotificationSettingsValidatorDashboard */
  settings?: any;
}

export interface TypesNotificationSettingsGeneral {
  /** notifications are disabled until this timestamp */
  do_not_disturb_timestamp?: number;
  is_email_notifications_enabled?: boolean;
  is_machine_cpu_usage_subscribed?: boolean;
  is_machine_memory_usage_subscribed?: boolean;
  is_machine_offline_subscribed?: boolean;
  is_machine_storage_usage_subscribed?: boolean;
  is_push_notifications_enabled?: boolean;
  is_webhook_notifications_enabled?: boolean;
  machine_cpu_usage_threshold?: number;
  machine_memory_usage_threshold?: number;
  machine_storage_usage_threshold?: number;
}

export interface TypesNotificationSettingsNetwork {
  gas_above_threshold?: number;
  gas_below_threshold?: number;
  is_gas_above_subscribed?: boolean;
  is_gas_below_subscribed?: boolean;
  is_new_reward_round_subscribed?: boolean;
  is_participation_rate_subscribed?: boolean;
  participation_rate_threshold?: number;
}

export interface TypesNotificationSettingsValidatorDashboard {
  group_efficiency_below_threshold?: number;
  is_attestations_missed_subscribed?: boolean;
  is_block_proposal_missed_subscribed?: boolean;
  is_block_proposal_success_subscribed?: boolean;
  is_group_efficiency_below_subscribed?: boolean;
  is_max_collateral_subscribed?: boolean;
  is_min_collateral_subscribed?: boolean;
  is_slashed_subscribed?: boolean;
  is_sync_subscribed?: boolean;
  is_upcoming_block_proposal_subscribed?: boolean;
  is_validator_offline_subscribed?: boolean;
  is_webhook_discord_enabled?: boolean;
  is_withdrawal_processed_subscribed?: boolean;
  max_collateral_threshold?: number;
  min_collateral_threshold?: number;
  webhook_url?: string;
}

export interface TypesNotificationValidatorDashboardDetail {
  /** index (epoch) */
  attestation_missed?: TypesIndexEpoch[];
  dashboard_name?: string;
  /** fill with the `group_efficiency_below` threshold if event is present */
  group_efficiency_below?: number;
  group_name?: string;
  /** node addresses */
  max_collateral?: TypesAddress[];
  /** node addresses */
  min_collateral?: TypesAddress[];
  proposal_missed?: TypesIndexSlots[];
  proposal_success?: TypesIndexBlocks[];
  proposal_upcoming?: TypesIndexSlots[];
  /** validator indices */
  slashed?: number[];
  /** validator indices */
  sync?: number[];
  /** validator indices */
  validator_offline?: number[];
  /** validator indices; TODO not filled yet */
  validator_offline_reminder?: number[];
  validator_online?: TypesNotificationEventValidatorBackOnline[];
  withdrawal?: TypesNotificationEventWithdrawal[];
}

export interface TypesPaging {
  next_cursor?: string;
  prev_cursor?: string;
  total_count?: number;
}

export interface TypesPercentageDetailsDecimalDecimal {
  max_value?: number;
  min_value?: number;
  percentage?: number;
}

export interface TypesPeriodicValuesFloat64 {
  all_time?: number;
  last_24h?: number;
  last_30d?: number;
  last_7d?: number;
}

export interface TypesPeriodicValuesTypesClElValueDecimalDecimal {
  all_time?: TypesClElValueDecimalDecimal;
  last_24h?: TypesClElValueDecimalDecimal;
  last_30d?: TypesClElValueDecimalDecimal;
  last_7d?: TypesClElValueDecimalDecimal;
}

export interface TypesPeriodicValuesTypesClElValueFloat64 {
  all_time?: TypesClElValueFloat64;
  last_24h?: TypesClElValueFloat64;
  last_30d?: TypesClElValueFloat64;
  last_7d?: TypesClElValueFloat64;
}

export interface TypesPostValidatorDashboardGroupsRequest {
  name?: string;
}

export interface TypesPutUserNotificationSettingsAccountDashboardResponse {
  data?: TypesNotificationSettingsAccountDashboard;
}

export interface TypesPutUserNotificationSettingsClientResponse {
  data?: TypesNotificationSettingsClient;
}

export interface TypesPutUserNotificationSettingsGeneralResponse {
  data?: TypesNotificationSettingsGeneral;
}

export interface TypesPutUserNotificationSettingsNetworksResponse {
  data?: TypesNotificationNetwork;
}

export interface TypesPutUserNotificationSettingsPairedDevicesResponse {
  data?: TypesNotificationPairedDevice;
}

export interface TypesPutUserNotificationSettingsValidatorDashboardResponse {
  data?: TypesNotificationSettingsValidatorDashboard;
}

export interface TypesSlotVizEpoch {
  epoch?: number;
  /** only on landing page */
  progress?: number;
  /** only on dashboard page */
  slots?: TypesVDBSlotVizSlot[];
  /** all on landing page, only 'head' on dashboard page */
  state?: string;
}

export interface TypesStatusCount {
  failed?: number;
  success?: number;
}

export interface TypesUserDashboardsData {
  account_dashboards?: TypesAccountDashboard[];
  validator_dashboards?: TypesValidatorDashboard[];
}

export interface TypesVDBBlocksTableRow {
  proposer?: number;
  group_id?: number;
  epoch?: number;
  slot?: number;
  block?: number;
  graffiti?: string;
  reward?: TypesClElValueDecimalDecimal;
  reward_recipient?: TypesAddress;
  status?: string;
}

export interface TypesVDBConsensusDepositsTableRow {
  amount?: number;
  epoch?: number;
  group_id?: number;
  index?: number;
  public_key?: string;
  signature?: string;
  slot?: number;
  withdrawal_credential?: string;
}

export interface TypesVDBEpochDutiesTableRow {
  validator?: number;
  duties?: TypesValidatorHistoryDuties;
}

export interface TypesVDBExecutionDepositsTableRow {
  amount?: number;
  block?: number;
  depositor?: TypesAddress;
  from?: TypesAddress;
  group_id?: number;
  index?: number;
  public_key?: string;
  timestamp?: number;
  tx_hash?: string;
  valid?: boolean;
  withdrawal_credential?: string;
}

export interface TypesVDBGroupRewardsData {
  attestations_head?: TypesVDBGroupRewardsDetails;
  attestations_source?: TypesVDBGroupRewardsDetails;
  attestations_target?: TypesVDBGroupRewardsDetails;
  inactivity?: TypesVDBGroupRewardsDetails;
  proposal_cl_att_inc_reward?: number;
  proposal_cl_slashing_inc_reward?: number;
  proposal_cl_sync_inc_reward?: number;
  proposal_el_reward?: number;
  proposal_status_count?: TypesStatusCount;
  slashing?: TypesVDBGroupRewardsDetails;
  sync?: TypesVDBGroupRewardsDetails;
}

export interface TypesVDBGroupRewardsDetails {
  income?: number;
  status_count?: TypesStatusCount;
}

export interface TypesVDBGroupSummaryColumnItem {
  status_count?: TypesStatusCount;
  /** number of distinct validators */
  validator_count?: number;
  /** fill with up to 3 validator indexes */
  validators?: number[];
}

export interface TypesVDBGroupSummaryData {
  apr?: TypesClElValueFloat64;
  attestation_avg_incl_dist?: number;
  attestation_efficiency?: number;
  attestations_head?: TypesStatusCount;
  attestations_source?: TypesStatusCount;
  attestations_target?: TypesStatusCount;
  balances?: TypesValidatorBalances;
  efficiency?: number;
  luck?: TypesLuck;
  missed_rewards?: TypesVDBGroupSummaryMissedRewards;
  /** number of distinct validators */
  proposal_validator_count?: number;
  /** fill with up to 3 validator indexes */
  proposal_validators?: number[];
  rewards?: TypesClElValueDecimalDecimal;
  rocket_pool?: {
    collateral?: number;
    minipools?: number;
  };
  /** Failed slashings are count of validators in the group that were slashed */
  slashings?: TypesVDBGroupSummaryColumnItem;
  sync?: TypesVDBGroupSummaryColumnItem;
  sync_count?: TypesVDBGroupSummarySyncCount;
}

export interface TypesVDBGroupSummaryMissedRewards {
  attestations?: number;
  proposer_rewards?: TypesClElValueDecimalDecimal;
  sync?: number;
}

export interface TypesVDBGroupSummarySyncCount {
  current_validators?: number;
  past_periods?: number;
  upcoming_validators?: number;
}

export interface TypesVDBManageValidatorsTableRow {
  balance?: number;
  group_id?: number;
  index?: number;
  public_key?: string;
  status?: string;
  withdrawal_credential?: string;
}

export interface TypesVDBOverviewData {
  name?: string;
  apr?: TypesPeriodicValuesTypesClElValueFloat64;
  balances?: TypesValidatorBalances;
  chart_history_seconds?: TypesChartHistorySeconds;
  efficiency?: TypesPeriodicValuesFloat64;
  groups?: TypesVDBOverviewGroup[];
  network?: number;
  rewards?: TypesPeriodicValuesTypesClElValueDecimalDecimal;
  validators?: TypesValidatorStateCounts;
}

export interface TypesVDBOverviewGroup {
  count?: number;
  id?: number;
  name?: string;
}

export interface TypesVDBPostArchivingReturnData {
  id?: number;
  is_archived?: boolean;
}

export interface TypesVDBPostCreateGroupData {
  id?: number;
  name?: string;
}

export interface TypesVDBPostReturnData {
  created_at?: number;
  id?: number;
  name?: string;
  network?: number;
  user_id?: number;
}

export interface TypesVDBPostValidatorsData {
  group_id?: number;
  index?: number;
}

export interface TypesVDBPublicId {
  name?: string;
  public_id?: string;
  share_settings?: {
    share_groups?: boolean;
  };
}

export interface TypesVDBRewardsTableDuty {
  attestation?: number;
  proposal?: number;
  slashing?: number;
  sync?: number;
}

export interface TypesVDBRewardsTableRow {
  duty?: TypesVDBRewardsTableDuty;
  epoch?: number;
  group_id?: number;
  reward?: TypesClElValueDecimalDecimal;
}

export interface TypesVDBRocketPoolMinipoolsTableRow {
  commission?: number;
  created_timestamp?: number;
  deposit?: number;
  group_id?: number;
  minipool_status?: string;
  node?: TypesAddress;
  penalties?: number;
  validator_index?: number;
  validator_status?: string;
}

export interface TypesVDBRocketPoolTableRow {
  node?: TypesAddress;
  avg_commission?: number;
  collateral?: TypesPercentageDetailsDecimalDecimal;
  deposit_credit?: number;
  effective_rpl?: number;
  minipools_count_leb_16?: number;
  minipools_count_leb_8?: number;
  minipools_count_total?: number;
  node_deposit_balance?: number;
  refund_balance?: number;
  rpl_apr?: number;
  rpl_apr_update_ts?: number;
  rpl_claimed?: number;
  rpl_estimate?: number;
  rpl_unclaimed?: number;
  smoothingpool_claimed?: number;
  smoothingpool_opt_in?: boolean;
  smoothingpool_unclaimed?: number;
  staked_eth?: number;
  staked_rpl?: number;
  timezone?: string;
  user_deposit_balance?: number;
}

export interface TypesVDBSlotVizDuty {
  total_count?: number;
  /** up to 6 validators that performed the duty, only for scheduled and failed */
  validators?: number[];
}

export interface TypesVDBSlotVizSlashing {
  /** up to 6 slashings, validator is always the slashing validator */
  slashings?: TypesVDBSlotVizTuple[];
  total_count?: number;
}

export interface TypesVDBSlotVizSlot {
  attestations?: TypesVDBSlotVizStatusTypesVDBSlotVizDuty;
  proposal?: TypesVDBSlotVizTuple;
  slashing?: TypesVDBSlotVizStatusTypesVDBSlotVizSlashing;
  slot?: number;
  status?: string;
  sync?: TypesVDBSlotVizStatusTypesVDBSlotVizDuty;
}

export interface TypesVDBSlotVizStatusTypesVDBSlotVizDuty {
  failed?: TypesVDBSlotVizDuty;
  scheduled?: TypesVDBSlotVizDuty;
  success?: TypesVDBSlotVizDuty;
}

export interface TypesVDBSlotVizStatusTypesVDBSlotVizSlashing {
  failed?: TypesVDBSlotVizSlashing;
  scheduled?: TypesVDBSlotVizSlashing;
  success?: TypesVDBSlotVizSlashing;
}

export interface TypesVDBSlotVizTuple {
  /**
   * If the duty is a proposal & it's successful, the duty_object is the proposed block
   * If the duty is a proposal & it failed/scheduled, the duty_object is the slot
   * If the duty is a slashing & it's successful, the duty_object is the validator you slashed
   * If the duty is a slashing & it failed, the duty_object is your validator that was slashed
   */
  duty_object?: number;
  validator?: number;
}

export interface TypesVDBSummaryStatus {
  current_sync_count?: number;
  next_sync_count?: number;
  slashed_count?: number;
}

export interface TypesVDBSummaryTableRow {
  group_id?: number;
  attestations?: TypesStatusCount;
  average_network_efficiency?: number;
  efficiency?: number;
  proposals?: TypesStatusCount;
  reward?: TypesClElValueDecimalDecimal;
  status?: TypesVDBSummaryStatus;
  validators?: TypesVDBSummaryValidators;
}

export interface TypesVDBSummaryValidator {
  index?: number;
  duty_objects?: number[];
}

export interface TypesVDBSummaryValidators {
  exited?: number;
  offline?: number;
  online?: number;
}

export interface TypesVDBSummaryValidatorsData {
  category?: string;
  validators?: TypesVDBSummaryValidator[];
}

export interface TypesVDBTotalConsensusDepositsData {
  total_amount?: number;
}

export interface TypesVDBTotalExecutionDepositsData {
  total_amount?: number;
}

export interface TypesVDBTotalWithdrawalsData {
  total_amount?: number;
}

export interface TypesVDBWithdrawalsTableRow {
  amount?: number;
  epoch?: number;
  group_id?: number;
  index?: number;
  is_missing_estimate?: boolean;
  recipient?: TypesAddress;
  slot?: number;
}

export interface TypesValidatorBalances {
  /** on-chain */
  effective_current?: number;
  /** from premium perks pov: exited validators are counted with their latest eb */
  effective_latest?: number;
  staked_eth?: number;
  total?: number;
}

export interface TypesValidatorDashboard {
  id?: number;
  name?: string;
  network?: number;
  public_ids?: TypesVDBPublicId[];
  is_archived?: boolean;
  archived_reason?: string;
  validator_count?: number;
  group_count?: number;
}

export interface TypesValidatorHistoryDuties {
  attestation_head?: TypesValidatorHistoryEvent;
  attestation_source?: TypesValidatorHistoryEvent;
  attestation_target?: TypesValidatorHistoryEvent;
  proposal?: TypesValidatorHistoryProposal;
  slashing?: TypesValidatorHistoryEvent;
  sync?: TypesValidatorHistoryEvent;
  /** count of successful sync duties for the epoch */
  sync_count?: number;
}

export interface TypesValidatorHistoryEvent {
  income?: number;
  status?: string;
}

export interface TypesValidatorHistoryProposal {
  cl_attestation_inclusion_income?: number;
  cl_slashing_inclusion_income?: number;
  cl_sync_inclusion_income?: number;
  el_income?: number;
  status?: string;
}

export interface TypesValidatorStateCounts {
  exited?: number;
  offline?: number;
  online?: number;
  pending?: number;
  slashed?: number;
}

export type QueryParamsType = Record<string | number, any>;
export type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;

export interface FullRequestParams extends Omit<RequestInit, "body"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseFormat;
  /** request body */
  body?: unknown;
  /** base url */
  baseUrl?: string;
  /** request cancellation token */
  cancelToken?: CancelToken;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown> {
  baseUrl?: string;
  baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<RequestParams | void> | RequestParams | void;
  customFetch?: typeof fetch;
}

export interface HttpResponse<D extends unknown, E extends unknown = unknown>
  extends Response {
  data: D;
  error: E;
}

type CancelToken = Symbol | string | number;

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public baseUrl: string = "/api/v2";
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private abortControllers = new Map<CancelToken, AbortController>();
  private customFetch = (...fetchParams: Parameters<typeof fetch>) =>
    fetch(...fetchParams);

  private baseApiParams: RequestParams = {
    credentials: "same-origin",
    headers: {},
    redirect: "follow",
    referrerPolicy: "no-referrer",
  };

  constructor(apiConfig: ApiConfig<SecurityDataType> = {}) {
    Object.assign(this, apiConfig);
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected encodeQueryParam(key: string, value: any) {
    const encodedKey = encodeURIComponent(key);
    return `${encodedKey}=${encodeURIComponent(typeof value === "number" ? value : `${value}`)}`;
  }

  protected addQueryParam(query: QueryParamsType, key: string) {
    return this.encodeQueryParam(key, query[key]);
  }

  protected addArrayQueryParam(query: QueryParamsType, key: string) {
    const value = query[key];
    return value.map((v: any) => this.encodeQueryParam(key, v)).join("&");
  }

  protected toQueryString(rawQuery?: QueryParamsType): string {
    const query = rawQuery || {};
    const keys = Object.keys(query).filter(
      (key) => "undefined" !== typeof query[key],
    );
    return keys
      .map((key) =>
        Array.isArray(query[key])
          ? this.addArrayQueryParam(query, key)
          : this.addQueryParam(query, key),
      )
      .join("&");
  }

  protected addQueryParams(rawQuery?: QueryParamsType): string {
    const queryString = this.toQueryString(rawQuery);
    return queryString ? `?${queryString}` : "";
  }

  private contentFormatters: Record<ContentType, (input: any) => any> = {
    [ContentType.Json]: (input: any) =>
      input !== null && (typeof input === "object" || typeof input === "string")
        ? JSON.stringify(input)
        : input,
    [ContentType.Text]: (input: any) =>
      input !== null && typeof input !== "string"
        ? JSON.stringify(input)
        : input,
    [ContentType.FormData]: (input: any) =>
      Object.keys(input || {}).reduce((formData, key) => {
        const property = input[key];
        formData.append(
          key,
          property instanceof Blob
            ? property
            : typeof property === "object" && property !== null
              ? JSON.stringify(property)
              : `${property}`,
        );
        return formData;
      }, new FormData()),
    [ContentType.UrlEncoded]: (input: any) => this.toQueryString(input),
  };

  protected mergeRequestParams(
    params1: RequestParams,
    params2?: RequestParams,
  ): RequestParams {
    return {
      ...this.baseApiParams,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...(this.baseApiParams.headers || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected createAbortSignal = (
    cancelToken: CancelToken,
  ): AbortSignal | undefined => {
    if (this.abortControllers.has(cancelToken)) {
      const abortController = this.abortControllers.get(cancelToken);
      if (abortController) {
        return abortController.signal;
      }
      return void 0;
    }

    const abortController = new AbortController();
    this.abortControllers.set(cancelToken, abortController);
    return abortController.signal;
  };

  public abortRequest = (cancelToken: CancelToken) => {
    const abortController = this.abortControllers.get(cancelToken);

    if (abortController) {
      abortController.abort();
      this.abortControllers.delete(cancelToken);
    }
  };

  public request = async <T = any, E = any>({
    body,
    secure,
    path,
    type,
    query,
    format,
    baseUrl,
    cancelToken,
    ...params
  }: FullRequestParams): Promise<HttpResponse<T, E>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.baseApiParams.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const queryString = query && this.toQueryString(query);
    const payloadFormatter = this.contentFormatters[type || ContentType.Json];
    const responseFormat = format || requestParams.format;

    return this.customFetch(
      `${baseUrl || this.baseUrl || ""}${path}${queryString ? `?${queryString}` : ""}`,
      {
        ...requestParams,
        headers: {
          ...(requestParams.headers || {}),
          ...(type && type !== ContentType.FormData
            ? { "Content-Type": type }
            : {}),
        },
        signal:
          (cancelToken
            ? this.createAbortSignal(cancelToken)
            : requestParams.signal) || null,
        body:
          typeof body === "undefined" || body === null
            ? null
            : payloadFormatter(body),
      },
    ).then(async (response) => {
      const r = response.clone() as HttpResponse<T, E>;
      r.data = null as unknown as T;
      r.error = null as unknown as E;

      const data = !responseFormat
        ? r
        : await response[responseFormat]()
            .then((data) => {
              if (r.ok) {
                r.data = data;
              } else {
                r.error = data;
              }
              return r;
            })
            .catch((e) => {
              r.error = e;
              return r;
            });

      if (cancelToken) {
        this.abortControllers.delete(cancelToken);
      }

      if (!response.ok) throw data;
      return data;
    });
  };
}

/**
 * @title beaconcha.in API
 * @version 2.0
 * @baseUrl /api/v2
 * @contact
 *
 * To authenticate your API request beaconcha.in uses API Keys. Set your API Key either by:
 * - Setting the `Authorization` header in the following format: `Authorization: Bearer <your-api-key>`. (recommended)
 * - Setting the URL query parameter in the following format: `api_key={your_api_key}`.\
 * Example: `https://beaconcha.in/api/v2/example?field=value&api_key={your_api_key}`
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  users = {
    /**
     * @description Get all dashboards of the authenticated user.
     *
     * @tags Dashboards
     * @name MeDashboardsList
     * @request GET:/users/me/dashboards
     * @secure
     */
    meDashboardsList: (params: RequestParams = {}) =>
      this.request<TypesApiDataResponseTypesUserDashboardsData, any>({
        path: `/users/me/dashboards`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get an overview of your recent notifications.
     *
     * @tags Notifications
     * @name MeNotificationsList
     * @request GET:/users/me/notifications
     * @secure
     */
    meNotificationsList: (params: RequestParams = {}) =>
      this.request<TypesGetUserNotificationsResponse, any>({
        path: `/users/me/notifications`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a detailed view of a triggered notification related to an account dashboard group at a specific epoch.
     *
     * @tags Notifications
     * @name MeNotificationsAccountDashboardsGroupsEpochsDetail
     * @request GET:/users/me/notifications/account-dashboards/{dashboard_id}/groups/{group_id}/epochs/{epoch}
     * @secure
     */
    meNotificationsAccountDashboardsGroupsEpochsDetail: (
      dashboardId: string,
      groupId: number,
      epoch: number,
      query?: {
        /** Search for Address, ENS */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationsAccountDashboardResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/account-dashboards/${dashboardId}/groups/${groupId}/epochs/${epoch}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of triggered notifications related to your clients.
     *
     * @tags Notifications
     * @name MeNotificationsClientsList
     * @request GET:/users/me/notifications/clients
     * @secure
     */
    meNotificationsClientsList: (
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: number;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "client_name" | "timestamp";
        /** Search for Client */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationClientsResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/clients`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of triggered notifications related to your dashboards.
     *
     * @tags Notifications
     * @name MeNotificationsDashboardsList
     * @request GET:/users/me/notifications/dashboards
     * @secure
     */
    meNotificationsDashboardsList: (
      query?: {
        /** If set, results will be filtered to only include networks given. Provide a comma separated list. */
        networks?: string;
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: number;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "chain_id" | "timestamp" | "dashboard_id";
        /** Search for Dashboard, Group */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationDashboardsResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/dashboards`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of triggered notifications related to your machines.
     *
     * @tags Notifications
     * @name MeNotificationsMachinesList
     * @request GET:/users/me/notifications/machines
     * @secure
     */
    meNotificationsMachinesList: (
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: number;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "machine_name" | "threshold" | "event_type" | "timestamp";
        /** Search for Machine */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationMachinesResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/machines`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of triggered notifications related to networks.
     *
     * @tags Notifications
     * @name MeNotificationsNetworksList
     * @request GET:/users/me/notifications/networks
     * @secure
     */
    meNotificationsNetworksList: (
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: number;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "timestamp" | "event_type";
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationNetworksResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/networks`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get notification settings for the authenticated user. Excludes dashboard notification settings.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsList
     * @request GET:/users/me/notifications/settings
     * @secure
     */
    meNotificationsSettingsList: (params: RequestParams = {}) =>
      this.request<TypesGetUserNotificationSettingsResponse, any>({
        path: `/users/me/notifications/settings`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update the notification settings for a specific group of an account dashboard for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsAccountDashboardsGroupsUpdate
     * @request PUT:/users/me/notifications/settings/account-dashboards/{dashboard_id}/groups/{group_id}
     * @secure
     */
    meNotificationsSettingsAccountDashboardsGroupsUpdate: (
      dashboardId: string,
      groupId: number,
      request: HandlersPublicPutUserNotificationSettingsAccountDashboardRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsAccountDashboardResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/account-dashboards/${dashboardId}/groups/${groupId}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update client notification settings for the authenticated user. When a client is subscribed, notifications will be sent when a new version is available.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsClientsUpdate
     * @request PUT:/users/me/notifications/settings/clients/{client_id}
     * @secure
     */
    meNotificationsSettingsClientsUpdate: (
      clientId: number,
      request: HandlersPublicPutUserNotificationSettingsClientRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsClientResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/clients/${clientId}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of notification settings for the dashboards of the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsDashboardsList
     * @request GET:/users/me/notifications/settings/dashboards
     * @secure
     */
    meNotificationsSettingsDashboardsList: (
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: number;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: string;
        /** Search for Dashboard, Group */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationSettingsDashboardsResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/dashboards`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update general notification settings for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsGeneralUpdate
     * @request PUT:/users/me/notifications/settings/general
     * @secure
     */
    meNotificationsSettingsGeneralUpdate: (
      request: TypesNotificationSettingsGeneral,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsGeneralResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/general`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update network notification settings for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsNetworksUpdate
     * @request PUT:/users/me/notifications/settings/networks/{network}
     * @secure
     */
    meNotificationsSettingsNetworksUpdate: (
      network: string,
      request: HandlersPublicPutUserNotificationSettingsNetworksRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsNetworksResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/networks/${network}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update paired device notification settings for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsPairedDevicesUpdate
     * @request PUT:/users/me/notifications/settings/paired-devices/{paired_device_id}
     * @secure
     */
    meNotificationsSettingsPairedDevicesUpdate: (
      pairedDeviceId: string,
      request: HandlersPublicPutUserNotificationSettingsPairedDevicesRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsPairedDevicesResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/paired-devices/${pairedDeviceId}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete paired device notification settings for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsPairedDevicesDelete
     * @request DELETE:/users/me/notifications/settings/paired-devices/{paired_device_id}
     * @secure
     */
    meNotificationsSettingsPairedDevicesDelete: (
      pairedDeviceId: string,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/users/me/notifications/settings/paired-devices/${pairedDeviceId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Update the notification settings for a specific group of a validator dashboard for the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsSettingsValidatorDashboardsGroupsUpdate
     * @request PUT:/users/me/notifications/settings/validator-dashboards/{dashboard_id}/groups/{group_id}
     * @secure
     */
    meNotificationsSettingsValidatorDashboardsGroupsUpdate: (
      dashboardId: string,
      groupId: number,
      request: TypesNotificationSettingsValidatorDashboard,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesPutUserNotificationSettingsValidatorDashboardResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/settings/validator-dashboards/${dashboardId}/groups/${groupId}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Send a test email notification to the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsTestEmailCreate
     * @request POST:/users/me/notifications/test-email
     * @secure
     */
    meNotificationsTestEmailCreate: (params: RequestParams = {}) =>
      this.request<void, any>({
        path: `/users/me/notifications/test-email`,
        method: "POST",
        secure: true,
        ...params,
      }),

    /**
     * @description Send a test push notification to the authenticated user.
     *
     * @tags Notification Settings
     * @name MeNotificationsTestPushCreate
     * @request POST:/users/me/notifications/test-push
     * @secure
     */
    meNotificationsTestPushCreate: (params: RequestParams = {}) =>
      this.request<void, any>({
        path: `/users/me/notifications/test-push`,
        method: "POST",
        secure: true,
        ...params,
      }),

    /**
     * @description Send a test webhook notification from the authenticated user to the given URL.
     *
     * @tags Notification Settings
     * @name MeNotificationsTestWebhookCreate
     * @request POST:/users/me/notifications/test-webhook
     * @secure
     */
    meNotificationsTestWebhookCreate: (
      request: HandlersPublicPostUserNotificationsTestWebhookRequest,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/users/me/notifications/test-webhook`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * @description Get a detailed view of a triggered notification related to a validator dashboard group at a specific epoch.
     *
     * @tags Notifications
     * @name MeNotificationsValidatorDashboardsGroupsEpochsDetail
     * @request GET:/users/me/notifications/validator-dashboards/{dashboard_id}/groups/{group_id}/epochs/{epoch}
     * @secure
     */
    meNotificationsValidatorDashboardsGroupsEpochsDetail: (
      dashboardId: string,
      groupId: number,
      epoch: number,
      query?: {
        /** Search for Index */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetUserNotificationsValidatorDashboardResponse,
        TypesApiErrorResponse
      >({
        path: `/users/me/notifications/validator-dashboards/${dashboardId}/groups/${groupId}/epochs/${epoch}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),
  };
  validatorDashboards = {
    /**
     * @description Create a new validator dashboard. **Note**: New dashboards will automatically have a default group created.
     *
     * @tags Validator Dashboard Management
     * @name ValidatorDashboardsCreate
     * @request POST:/validator-dashboards
     * @secure
     */
    validatorDashboardsCreate: (
      request: HandlersPublicPostValidatorDashboardsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseTypesVDBPostReturnData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get overview information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name ValidatorDashboardsDetail
     * @request GET:/validator-dashboards/{dashboard_id}
     */
    validatorDashboardsDetail: (
      dashboardId: string,
      query?: {
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<TypesGetValidatorDashboardResponse, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name ValidatorDashboardsDelete
     * @request DELETE:/validator-dashboards/{dashboard_id}
     * @secure
     */
    validatorDashboardsDelete: (
      dashboardId: number,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Archive or unarchive a specified validator dashboard. Archived dashboards cannot be accessed by other endpoints. Archiving happens automatically if the number of dashboards, validators, or groups exceeds the limit allowed by your subscription plan. For example, this might occur if you downgrade your subscription to a lower tier.
     *
     * @tags Validator Dashboard Management
     * @name ArchivingUpdate
     * @request PUT:/validator-dashboards/{dashboard_id}/archiving
     * @secure
     */
    archivingUpdate: (
      dashboardId: number,
      request: HandlersPublicPutValidatorDashboardArchivingRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseTypesVDBPostArchivingReturnData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/archiving`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get blocks information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name BlocksList
     * @request GET:/validator-dashboards/{dashboard_id}/blocks
     */
    blocksList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "proposer" | "slot" | "block" | "status" | "reward";
        /** Search for Index, Public Key, Group. */
        search?: string;
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardBlocksResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/blocks`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get consensus layer deposits information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name ConsensusLayerDepositsList
     * @request GET:/validator-dashboards/{dashboard_id}/consensus-layer-deposits
     */
    consensusLayerDepositsList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardConsensusLayerDepositsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/consensus-layer-deposits`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get duties information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name DutiesDetail
     * @request GET:/validator-dashboards/{dashboard_id}/duties/{epoch}
     */
    dutiesDetail: (
      dashboardId: string,
      epoch: number,
      query?: {
        /** The ID of the group. */
        group_id?: number;
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "validator" | "reward";
        /** Search for Index, Public Key. */
        search?: string;
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardDutiesResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/duties/${epoch}`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get execution layer deposits information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name ExecutionLayerDepositsList
     * @request GET:/validator-dashboards/{dashboard_id}/execution-layer-deposits
     */
    executionLayerDepositsList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardExecutionLayerDepositsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/execution-layer-deposits`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new group in a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name GroupsCreate
     * @request POST:/validator-dashboards/{dashboard_id}/groups
     * @secure
     */
    groupsCreate: (
      dashboardId: number,
      request: TypesPostValidatorDashboardGroupsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseTypesVDBPostCreateGroupData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/groups`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a groups name in a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name GroupsUpdate
     * @request PUT:/validator-dashboards/{dashboard_id}/groups/{group_id}
     * @secure
     */
    groupsUpdate: (
      dashboardId: number,
      groupId: number,
      request: HandlersPublicPutValidatorDashboardGroupsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseTypesVDBPostCreateGroupData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/groups/${groupId}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a group in a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name GroupsDelete
     * @request DELETE:/validator-dashboards/{dashboard_id}/groups/{group_id}
     * @secure
     */
    groupsDelete: (
      dashboardId: number,
      groupId: number,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}/groups/${groupId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Get rewards information for a specified group in a specified dashboard
     *
     * @tags Validator Dashboard
     * @name GroupsRewardsDetail
     * @request GET:/validator-dashboards/{dashboard_id}/groups/{group_id}/rewards/{epoch}
     */
    groupsRewardsDetail: (
      dashboardId: string,
      groupId: number,
      epoch: number,
      query?: {
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardGroupRewardsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/groups/${groupId}/rewards/${epoch}`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get summary information for a specified group in a specified dashboard
     *
     * @tags Validator Dashboard
     * @name GroupsSummaryList
     * @request GET:/validator-dashboards/{dashboard_id}/groups/{group_id}/summary
     */
    groupsSummaryList: (
      dashboardId: string,
      groupId: number,
      query: {
        /** Time period to get data for. */
        period: "all_time" | "last_30d" | "last_7d" | "last_24h" | "last_1h";
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardGroupSummaryResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/groups/${groupId}/summary`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete all validators from a specified group in a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name GroupsValidatorsDelete
     * @request DELETE:/validator-dashboards/{dashboard_id}/groups/{group_id}/validators
     * @secure
     */
    groupsValidatorsDelete: (
      dashboardId: number,
      groupId: number,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}/groups/${groupId}/validators`,
        method: "DELETE",
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * @description Update the name of a specified validator dashboard.
     *
     * @tags Validator Dashboard Management
     * @name NameUpdate
     * @request PUT:/validator-dashboards/{dashboard_id}/name
     * @secure
     */
    nameUpdate: (
      dashboardId: number,
      request: HandlersPublicPutValidatorDashboardNameRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseTypesVDBPostReturnData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/name`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new public ID for a specified dashboard. This can be used as an ID by other users for non-modifying (i.e. GET) endpoints only. Currently limited to one per dashboard.
     *
     * @tags Validator Dashboard Management
     * @name PublicIdsCreate
     * @request POST:/validator-dashboards/{dashboard_id}/public-ids
     * @secure
     */
    publicIdsCreate: (
      dashboardId: number,
      request: HandlersPublicPostValidatorDashboardPublicIdsRequest,
      params: RequestParams = {},
    ) =>
      this.request<TypesApiDataResponseTypesVDBPublicId, TypesApiErrorResponse>(
        {
          path: `/validator-dashboards/${dashboardId}/public-ids`,
          method: "POST",
          body: request,
          secure: true,
          type: ContentType.Json,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Update a specified public ID for a specified dashboard.
     *
     * @tags Validator Dashboard Management
     * @name PublicIdsUpdate
     * @request PUT:/validator-dashboards/{dashboard_id}/public-ids/{public_id}
     * @secure
     */
    publicIdsUpdate: (
      dashboardId: number,
      publicId: string,
      request: HandlersPublicPutValidatorDashboardPublicIdRequest,
      params: RequestParams = {},
    ) =>
      this.request<TypesApiDataResponseTypesVDBPublicId, TypesApiErrorResponse>(
        {
          path: `/validator-dashboards/${dashboardId}/public-ids/${publicId}`,
          method: "PUT",
          body: request,
          secure: true,
          type: ContentType.Json,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Delete a specified public ID for a specified dashboard.
     *
     * @tags Validator Dashboard Management
     * @name PublicIdsDelete
     * @request DELETE:/validator-dashboards/{dashboard_id}/public-ids/{public_id}
     * @secure
     */
    publicIdsDelete: (
      dashboardId: number,
      publicId: string,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}/public-ids/${publicId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Get rewards information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name RewardsList
     * @request GET:/validator-dashboards/{dashboard_id}/rewards
     */
    rewardsList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "epoch";
        /** Search for Epoch, Index, Public Key, Group. */
        search?: string;
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardRewardsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/rewards`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get rewards chart data for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name RewardsChartList
     * @request GET:/validator-dashboards/{dashboard_id}/rewards-chart
     */
    rewardsChartList: (
      dashboardId: string,
      query?: {
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardRewardsChartResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/rewards-chart`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get an aggregated list of the Rocket Pool nodes details associated with a specified dashboard.
     *
     * @tags Validator Dashboard
     * @name RocketPoolList
     * @request GET:/validator-dashboards/{dashboard_id}/rocket-pool
     */
    rocketPoolList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?:
          | "node"
          | "minipools"
          | "collateral"
          | "rpl"
          | "effective_rpl"
          | "rpl_apr"
          | "smoothing_pool";
        /** Search for Node address. */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardRocketPoolResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/rocket-pool`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get minipools information for a specified Rocket Pool node associated with a specified dashboard.
     *
     * @tags Validator Dashboard
     * @name RocketPoolMinipoolsList
     * @request GET:/validator-dashboards/{dashboard_id}/rocket-pool/{node_address}/minipools
     */
    rocketPoolMinipoolsList: (
      dashboardId: string,
      nodeAddress: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "group_id";
        /** Search for Index, Node. */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardRocketPoolMinipoolsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/rocket-pool/${nodeAddress}/minipools`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get slot viz information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name SlotVizList
     * @request GET:/validator-dashboards/{dashboard_id}/slot-viz
     */
    slotVizList: (
      dashboardId: string,
      query?: {
        /** Provide a comma separated list of group IDs to filter the results by. If omitted, all groups will be included. */
        group_ids?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardSlotVizResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/slot-viz`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get summary information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name SummaryList
     * @request GET:/validator-dashboards/{dashboard_id}/summary
     */
    summaryList: (
      dashboardId: string,
      query: {
        /** Time period to get data for. */
        period: "all_time" | "last_30d" | "last_7d" | "last_24h" | "last_1h";
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?:
          | "group_id"
          | "validators"
          | "efficiency"
          | "attestations"
          | "proposals"
          | "reward";
        /** Search for Index, Public Key, Group. */
        search?: string;
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardSummaryResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/summary`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get summary chart data for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name SummaryChartList
     * @request GET:/validator-dashboards/{dashboard_id}/summary-chart
     */
    summaryChartList: (
      dashboardId: string,
      query?: {
        /** Provide a comma separated list of group IDs to filter the results by. */
        group_ids?: string;
        /** Efficiency type to get data for. */
        efficiency_type?: "all" | "attestation" | "sync" | "proposal";
        /**
         * Aggregation type to get data for.
         * @default "hourly"
         */
        aggregation?: "epoch" | "hourly" | "daily" | "weekly";
        /** Return data after this timestamp. */
        after_ts?: string;
        /** Return data before this timestamp. */
        before_ts?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardSummaryChartResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/summary-chart`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get summary information for validators in a specified dashboard
     *
     * @tags Validator Dashboard
     * @name SummaryValidatorsList
     * @request GET:/validator-dashboards/{dashboard_id}/summary/validators
     */
    summaryValidatorsList: (
      dashboardId: string,
      query: {
        /** The ID of the group. */
        group_id?: number;
        /**
         * Validator duty to get data for.
         * @default "none"
         */
        duty?: "none" | "sync" | "slashed" | "proposal";
        /** Time period to get data for. */
        period: "all_time" | "last_30d" | "last_7d" | "last_24h" | "last_1h";
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardSummaryValidatorsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/summary/validators`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get total consensus layer deposits information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name TotalConsensusLayerDepositsList
     * @request GET:/validator-dashboards/{dashboard_id}/total-consensus-layer-deposits
     */
    totalConsensusLayerDepositsList: (
      dashboardId: string,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardTotalConsensusDepositsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/total-consensus-layer-deposits`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * @description Get total execution layer deposits information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name TotalExecutionLayerDepositsList
     * @request GET:/validator-dashboards/{dashboard_id}/total-execution-layer-deposits
     */
    totalExecutionLayerDepositsList: (
      dashboardId: string,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardTotalExecutionDepositsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/total-execution-layer-deposits`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * @description Get a summary of all Rocket Pool nodes details associated with a specified dashboard.
     *
     * @tags Validator Dashboard
     * @name TotalRocketPoolList
     * @request GET:/validator-dashboards/{dashboard_id}/total-rocket-pool
     */
    totalRocketPoolList: (dashboardId: string, params: RequestParams = {}) =>
      this.request<
        TypesGetValidatorDashboardTotalRocketPoolResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/total-rocket-pool`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * @description Get total withdrawals information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name TotalWithdrawalsList
     * @request GET:/validator-dashboards/{dashboard_id}/total-withdrawals
     */
    totalWithdrawalsList: (
      dashboardId: string,
      query?: {
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardTotalWithdrawalsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/total-withdrawals`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a list of validators in a specified validator dashboard.
     *
     * @tags Validator Dashboard
     * @name ValidatorsList
     * @request GET:/validator-dashboards/{dashboard_id}/validators
     */
    validatorsList: (
      dashboardId: string,
      query?: {
        /** The ID of the group. */
        group_id?: number;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?:
          | "index"
          | "public_key"
          | "balance"
          | "status"
          | "withdrawal_credentials";
        /** Search for Address, ENS. */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardValidatorsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/validators`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Add new validators to a specified dashboard or update the group of already-added validators. This endpoint will add all possible validators or return an error if the subscription plan limits are exceeded. The response will contain a list of added validators.
     *
     * @tags Validator Dashboard Management
     * @name ValidatorsCreate
     * @request POST:/validator-dashboards/{dashboard_id}/validators
     * @secure
     */
    validatorsCreate: (
      dashboardId: number,
      request: HandlersPublicPostValidatorDashboardValidatorsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        TypesApiDataResponseArrayTypesVDBPostValidatorsData,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/validators`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Remove all validators from a specified dashboard.
     *
     * @tags Validator Dashboard Management
     * @name ValidatorsDelete
     * @request DELETE:/validator-dashboards/{dashboard_id}/validators
     * @secure
     */
    validatorsDelete: (dashboardId: number, params: RequestParams = {}) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}/validators`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Remove specific validators from a specified dashboard in bulk.
     *
     * @tags Validator Dashboard Management
     * @name ValidatorsBulkDeletionsCreate
     * @request POST:/validator-dashboards/{dashboard_id}/validators/bulk-deletions
     * @secure
     */
    validatorsBulkDeletionsCreate: (
      dashboardId: number,
      request: HandlersPublicPostValidatorDashboardValidatorBulkDeletionsRequest,
      params: RequestParams = {},
    ) =>
      this.request<void, TypesApiErrorResponse>({
        path: `/validator-dashboards/${dashboardId}/validators/bulk-deletions`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * @description Get withdrawals information for a specified dashboard
     *
     * @tags Validator Dashboard
     * @name WithdrawalsList
     * @request GET:/validator-dashboards/{dashboard_id}/withdrawals
     */
    withdrawalsList: (
      dashboardId: string,
      query?: {
        /** Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward. */
        cursor?: string;
        /** The maximum number of results that may be returned. */
        limit?: string;
        /** The field you want to sort by. Append with `:desc` for descending order. */
        sort?: "epoch" | "slot" | "index" | "recipient" | "amount";
        /** Search for Index, Public Key, Address. */
        search?: string;
        /** Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``. */
        modes?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        TypesGetValidatorDashboardWithdrawalsResponse,
        TypesApiErrorResponse
      >({
        path: `/validator-dashboards/${dashboardId}/withdrawals`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),
  };
}
