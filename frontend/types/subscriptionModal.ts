import { ChainIDs } from './network'

export type ValidatorSubscriptionState = {
  offlineValidator: boolean,
  offlineGroup?: number,
  missedAttestations: boolean,
  proposedBlock: boolean,
  upcomingProposal: boolean,
  syncCommittee: boolean,
  withdrawed: boolean,
  shlashed: boolean,
  realTime?: boolean
}

export type AccountSubscriptionState = {
  incoming: boolean,
  outgoing: boolean,
  erc20: number,
  erc721: boolean,
  erc1155: boolean,
  networks: ChainIDs[],
  ignoreSpam: boolean
}
