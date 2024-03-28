export function useNetwork () {
  // TODO: Replace hardcoded Ethereum Holesky values with real network information once network endpoint is available
  const tsForSlot0 = 1695902400
  const secondsPerSlot = 12
  const slotsPerEpoch = 32

  function epochToTs (epoch: number): number | undefined {
    if (epoch < 0) {
      return undefined
    }

    return tsForSlot0 + ((epoch * slotsPerEpoch) * secondsPerSlot)
  }

  function epochsPerDay (): number {
    return 24 * 60 * 60 / (slotsPerEpoch * secondsPerSlot)
  }

  return { epochToTs, epochsPerDay }
}
