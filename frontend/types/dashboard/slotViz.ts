export type SlotVizIcons = 'attestation' | 'head_attestation' | 'source_attestation' | 'target_attestation' | 'proposal' | 'slashing' | 'sync'

export type SlotVizValidatorDuties = {
  type: 'proposal' | 'slashing' | 'attestation' | 'sync'
  pendingCount?: number,
  successCount?: number,
  successEarning?: string, // wei
  failedCount?: number,
  failedEarnings?: string, // wei
  validator?: number // in case it's a proposal
}

export type SlotVizSlot = {
  id: number
  state: 'orphaned' | 'missed' | 'proposed' | 'scheduled'
  duties?: SlotVizValidatorDuties[] // in case of a dashboard slotviz, the duties of the dashboard's validators
}

export type SlotVizEpoch = {
  id: number
  state: 'head' | 'finalized' | 'scheduled'
  slots: SlotVizSlot[]
}

export type SlotVizData = {
  currentSlot: number
  epochs: SlotVizEpoch[]
}
