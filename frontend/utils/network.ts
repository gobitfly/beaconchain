// TODO: Replace hardcoded Ethereum Mainnet values with real network information once network endpoint is available
// This should als most likely become a composable later
const tsForSlot0 = 1606820423
const secondsPerSlot = 12
const slotsPerEpoch = 32

export function epochToTs (epoch: number): number | undefined {
  if (epoch < 0) {
    return undefined
  }

  return tsForSlot0 + ((epoch * slotsPerEpoch) * secondsPerSlot)
}

export function epochsPerDay (): number {
  return 24 * 60 * 60 / (slotsPerEpoch * secondsPerSlot)
}
