export type SlotVizValidatorDuties = {
  type: 'propsal' | 'slashing' | 'attestation' | 'sync'
  successCount?: number,
  successEarning?: string, // wei
  failedCount?: number,
  failedEarnings?: string, // wei
  validator?: number // in case it's a propsal
}

export type SlotVizSlot = {
  state: 'orphaned' | 'missed' | 'proposed' | 'scheduled'
  duties?: SlotVizValidatorDuties[] // in case of a dashboard slotviz, the duties of the dashboard's validators
}

export type SlotVizEpoch = {
  state: 'head' | 'finalized' | 'scheduled'
  slots: SlotVizSlot[]
}

export type SlotVizData = {
  currentSlot: number
  epochs: SlotVizEpoch[]
}
