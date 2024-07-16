import { ChainIDs } from '../network'

// TODO: import from '~/types/api/notifications.ts' once the corresponding PR is corrected and merged https://github.com/gobitfly/beaconchain/pull/573
export interface NotificationSettingsValidatorDashboard {
  is_validator_offline_subscribed: boolean;
  group_offline_threshold: number /* float64 */;
  is_attestations_missed_subscribed: boolean;
  is_block_proposal_subscribed: boolean;
  is_upcoming_block_proposal_subscribed: boolean;
  is_sync_subscribed: boolean;
  is_withdrawal_processed_subscribed: boolean;
  is_slashed_subscribed: boolean;
  is_real_time_mode_enabled: boolean;
}

// TODO: import from '~/types/api/notifications.ts' once the corresponding PR is corrected and merged https://github.com/gobitfly/beaconchain/pull/573
export interface NotificationSettingsAccountDashboard {
  is_incoming_transactions_subscribed: boolean;
  is_outgoing_transactions_subscribed: boolean;
  is_erc20_token_transfers_subscribed: boolean;
  erc20_token_transfers_threshold: number /* float64 */;
  is_erc721_token_transfers_subscribed: boolean;
  is_erc1155_token_transfers_subscribed: boolean;
  subscribed_chain_ids: number /* uint64 */[];
  is_ignore_spam_transactions_enabled: boolean;
}

export interface InternalEntry {
  type: 'binary' | 'amount' | 'percent' | 'networks'
  networks?: ChainIDs[]
  check?: boolean,
  num?: number
}

export type APIentry = boolean | number | undefined | null | ChainIDs[]
