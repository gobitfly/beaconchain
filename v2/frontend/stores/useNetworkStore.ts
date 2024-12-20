import * as networkTs from '~/types/network'

export function useNetworkStore() {
  const runTimeNetwork = Number(useRuntimeConfig().public.chainIdByDefault) as networkTs.ChainIDs
  const currentNetwork = computed(() => runTimeNetwork)
  const availableNetworks = computed(() => [ runTimeNetwork ])
  const networkInfo = computed(() => networkTs.ChainInfo[currentNetwork.value])

  function isNetworkDisabled(chainId: networkTs.ChainIDs): boolean {
    // TODO: return `false` for everything once we are ready
    return (
      !useRuntimeConfig().public.showInDevelopment
      && chainId !== currentNetwork.value
    )
  }

  function isMainNet(): boolean {
    return networkTs.isMainNet(currentNetwork.value)
  }

  function isL1(): boolean {
    return networkTs.isL1(currentNetwork.value)
  }

  function epochsPerDay(): number {
    return networkTs.epochsPerDay(currentNetwork.value)
  }

  function epochToTs(epoch: number): number | undefined {
    return networkTs.epochToTs(currentNetwork.value, epoch)
  }

  const secondsPerEpoch = computed(() => networkTs.secondsPerEpoch(currentNetwork.value))

  function slotToTs(slot: number): number | undefined {
    return networkTs.slotToTs(currentNetwork.value, slot)
  }

  function tsToSlot(ts: number): number {
    return networkTs.tsToSlot(currentNetwork.value, ts)
  }

  function slotToEpoch(slot: number): number {
    return networkTs.slotToEpoch(currentNetwork.value, slot)
  }

  function tsToEpoch(ts: number): number {
    return slotToEpoch(tsToSlot(ts))
  }

  return {
    availableNetworks,
    currentNetwork,
    epochsPerDay,
    epochToTs,
    isL1,
    isMainNet,
    isNetworkDisabled,
    networkInfo,
    secondsPerEpoch,
    slotToEpoch,
    slotToTs,
    tsToEpoch,
    tsToSlot,
  }
}
