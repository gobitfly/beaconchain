import { ChainIDs } from './network'

/** for an option of type `number`: to mean that no value exists in the DB yet, set it to NaN; to mean that it is has a value but it is deactivated, set it to a negative value (for example 10% becomes -10) */
export type ValidatorSubscriptionState = {
  offlineValidator: boolean,
  offlineGroup?: number,
  missedAttestations: boolean,
  proposedBlock: boolean,
  upcomingProposal: boolean,
  syncCommittee: boolean,
  withdrawn: boolean,
  slashed: boolean,
  realTime?: boolean
}

/** for an option of type `number`: to mean that no value exists in the DB yet, set it to NaN; to mean that it is has a value but it is deactivated, set it to a negative value (for example $50 becomes -50) */
export type AccountSubscriptionState = {
  incoming: boolean,
  outgoing: boolean,
  erc20: number,
  erc721: boolean,
  erc1155: boolean,
  networks: ChainIDs[],
  ignoreSpam: boolean
}

/** for the communication with the API  // TODO: write here the identifiers actually used by the API */
export const SubscriptionJSONfields: Record<keyof (ValidatorSubscriptionState&AccountSubscriptionState), string> = {
  offlineValidator: 'offline_validator',
  offlineGroup: 'offline_group',
  missedAttestations: 'missed_attestations',
  proposedBlock: 'proposed_block',
  upcomingProposal: 'upcoming_proposal',
  syncCommittee: 'sync_committee',
  withdrawn: 'withdrawn',
  slashed: 'slashed',
  realTime: 'real_time',
  incoming: 'incoming',
  outgoing: 'outgoing',
  erc20: 'erc20',
  erc721: 'erc721',
  erc1155: 'erc1155',
  networks: 'networks',
  ignoreSpam: 'ignore_spam'
}

/** for internal use */
export interface CheckboxAndNumber {
  check: boolean,
  num: number
}
