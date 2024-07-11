// TODO: import from '~/types/api/notifications.ts' once the corresponding PR is corrected and merged https://github.com/gobitfly/beaconchain/pull/573
export interface NotificationEventsValidatorDashboard {
  validator_offline: boolean;
  group_offline: number|null /* float64 */;
  attestations_missed: boolean;
  block_proposal: boolean;
  upcoming_block_proposal: boolean;
  sync: boolean;
  withdrawal_processed: boolean;
  slashed: boolean;
  realtime_mode: boolean;
}

// TODO: import from '~/types/api/notifications.ts' once the corresponding PR is corrected and merged https://github.com/gobitfly/beaconchain/pull/573
export interface NotificationEventsAccountDashboard {
  incoming_transactions: boolean;
  outgoing_transactions: boolean;
  track_erc20_token_transfers: number|null /* float64 */;
  track_erc721_token_transfers: boolean;
  track_erc1155_token_transfers: boolean;
  networks: number /* uint64 */[];
  ignore_spam_transactions: boolean;
}

/** for internal use */
export interface CheckboxAndNumber {
  check: boolean,
  num: number|null
}

export type InputRow = 'binary' | 'amount' | 'percent' | 'networks'
