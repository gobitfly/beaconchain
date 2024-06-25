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
  incomingTransfers: 
}
