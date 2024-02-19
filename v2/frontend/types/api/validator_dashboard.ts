import { Paging, Address, Luck, LuckItem, StatusCount, ClElValue, ClElValueFloat, PeriodicClElValues, PeriodicClElValuesFloat } from './common.ts';

/* Do not change, this code is generated from Golang structs */








export interface VDBOverviewEfficiency {
    total: number;
    attestation: number;
    proposal: number;
    sync: number;
}
export interface VDBOverviewValidators {
    total: number;
    active: number;
    pending: number;
    exited: number;
    slashed: number;
}
export interface VDBOverviewGroup {
    id: number;
    name: string;
}
export interface VDBOverviewResponse {
    groups: VDBOverviewGroup[];
    validators: VDBOverviewValidators;
    efficiency: VDBOverviewEfficiency;
    rewards: PeriodicClElValues;
    luck: Luck;
    apr: PeriodicClElValuesFloat;
}



export interface VDBSlotVizDuty {
    type: 'proposal' | 'attestation' | 'sync' | 'slashing';
    pending_count: number;
    success_count: number;
    success_earnings: string;
    failed_count: number;
    failed_earnings: string;
    block?: number;
    validator?: number;
}
export interface VDBSlotVizSlot {
    slot: number;
    status: 'proposed' | 'missed' | 'scheduled' | 'orphaned';
    duties: VDBSlotVizDuty[];
}
export interface VDBSlotVizEpoch {
    epoch: number;
    state: 'head' | 'finalized' | 'scheduled';
    slots: VDBSlotVizSlot[];
}
export interface VDBSlotVizResponse {
    data: VDBSlotVizEpoch[];
}



export interface VDBSummaryTableRow {
    group_id: number;
    efficiency_day: number;
    efficiency_week: number;
    efficiency_month: number;
    efficiency_total: number;
    validators: number[];
}

export interface VDBSummaryTableResponse {
    paging: Paging;
    data: VDBSummaryTableRow[];
}


export interface VDBGroupSummaryColumnItem {
    status_count: StatusCount;
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
    slashed: VDBGroupSummaryColumnItem;
    apr: ClElValueFloat;
    income: ClElValue;
    luck: Luck;
}
export interface VDBGroupSummary {
    details_day: VDBGroupSummaryColumn;
    details_week: VDBGroupSummaryColumn;
    details_month: VDBGroupSummaryColumn;
    details_total: VDBGroupSummaryColumn;
}
export interface VDBGroupSummaryResponse {
    data: VDBGroupSummary;
}



export interface VDBRewardsTableRow {
    epoch: number;
    group_id: number;
    reward: ClElValue;
}
export interface VDBRewardsTableResponse {
    paging: Paging;
    data: VDBRewardsTableRow[];
}

export interface VDBGroupRewardsDetails {
    status_count: StatusCount;
    income: string;
}
export interface VDBGroupRewards {
    attestation_source: VDBGroupRewardsDetails;
    attestation_target: VDBGroupRewardsDetails;
    attestation_head: VDBGroupRewardsDetails;
    sync: VDBGroupRewardsDetails;
    slashing: VDBGroupRewardsDetails;
    proposal: VDBGroupRewardsDetails;
    proposal_el_reward: string;
}
export interface VDBGroupRewardsResponse {
    data: VDBGroupRewards;
}


export interface VDBEpochDutiesItem {
    status: 'success' | 'partial' | 'failed' | 'orphaned';
    reward: string;
}
export interface VDBEpochDutiesTableRow {
    validator: number;
    attestation_source: VDBEpochDutiesItem;
    attestation_target: VDBEpochDutiesItem;
    attestation_head: VDBEpochDutiesItem;
    proposal: VDBEpochDutiesItem;
    sync: VDBEpochDutiesItem;
    slashing: VDBEpochDutiesItem;
}
export interface VDBEpochDutiesTableResponse {
    paging: Paging;
    data: VDBEpochDutiesTableRow[];
}


export interface VDBBlocksTableRow {
    proposer: number;
    group_id: number;
    epoch: number;
    slot: number;
    block: number;
    status: 'success' | 'missed' | 'orphaned' | 'scheduled';
    reward: ClElValue;
}
export interface VDBBlocksTableResponse {
    paging: Paging;
    data: VDBBlocksTableRow[];
}

export interface VDBHeatmapCell {
    /** Epoch */
    x: number;
    /** Group ID */
    y: number;
    /** Attestaton Rewards */
    value: number;
}
export interface VDBHeatmap {
    /** X-Axis Categories */
    epochs: number[];
    /** Y-Axis Categories */
    group_ids: number[];
    data: VDBHeatmapCell[];
}
export interface VDBHeatmapResponse {
    data: VDBHeatmap;
}


export interface VDBHeatmapTooltipDuty {
    validator: number;
    status: 'success' | 'failed' | 'orphaned';
}
export interface VDBHeatmapTooltipResponse {
    epoch: number;
    proposers: VDBHeatmapTooltipDuty[];
    syncs: VDBHeatmapTooltipDuty[];
    slashings: VDBHeatmapTooltipDuty[];
    attestation_head: StatusCount;
    attestation_source: StatusCount;
    attestation_target: StatusCount;
    attestation_income: string;
}


export interface VDBExecutionDepositsTableRow {
    public_key: string;
    index: number;
    group_id: number;
    block: number;
    from: Address;
    depositor: Address;
    tx_hash: string;
    withdrawal_credentials: string;
    amount: string;
    valid: boolean;
}
export interface VDBExecutionDepositsTableResponse {
    paging: Paging;
    data: VDBExecutionDepositsTableRow[];
}

export interface VDBConsensusDepositsTableRow {
    public_key: string;
    index: number;
    group_id: number;
    epoch: number;
    slot: number;
    withdrawal_credential: string;
    amount: string;
    signature: string;
}
export interface VDBConsensusDepositsTableResponse {
    paging: Paging;
    data: VDBConsensusDepositsTableRow[];
}

export interface VDBWithdrawalsTableRow {
    epoch: number;
    index: number;
    group_id: number;
    recipient: Address;
    amount: string;
}
export interface VDBWithdrawalsTableResponse {
    paging: Paging;
    data: VDBWithdrawalsTableRow[];
}

export interface VDBManageValidatorsTableRow {
    index: number;
    public_key: string;
    group_id: number;
    balance: string;
    status: string;
    withdrawal_credential: string;
}
export interface VDBManageValidatorsTableResponse {
    paging: Paging;
    data: VDBManageValidatorsTableRow[];
}
