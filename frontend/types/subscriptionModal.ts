import { ChainIDs } from './network'

/** translates our the names of our object members to/from the names used by the API  // TODO: write here the identifiers actually used by the API */
export enum SubscriptionJSONfields {
  offlineValidator = 'offline_validator',
  offline_validator = 'offlineValidator',

  offlineGroup = 'offline_group',
  offline_group = 'offlineGroup',

  missedAttestations = 'missed_attestations',
  missed_attestations = 'missedAttestations',

  proposedBlock = 'proposed_block',
  proposed_block = 'proposedBlock',

  upcomingProposal = 'upcoming_proposal',
  upcoming_proposal = 'upcomingProposal',

  syncCommittee = 'sync_committee',
  sync_committee = 'syncCommittee',

  withdrawn = 'withdrawn',

  slashed = 'slashed',

  realTime = 'real_time',
  real_time = 'realTime',

  incoming = 'incoming',

  outgoing = 'outgoing',

  erc20 = 'erc20',

  erc721 = 'erc721',

  erc1155 = 'erc1155',

  networks = 'networks',

  ignoreSpam = 'ignore_spam',
  ignore_spam = 'ignoreSpam'
}

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

/** for internal use */
export interface CheckboxAndNumber {
  check: boolean,
  num: number
}
