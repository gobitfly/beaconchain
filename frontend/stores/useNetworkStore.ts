import { defineStore } from 'pinia'

import {
  type ChainId,
  ChainIDs,
  ChainInfo,
} from '~/types/network'

const store = defineStore('network-store', () => {
  const data = ref<{
    availableNetworks: ChainId[],
    currentNetwork: ChainId,
  }>({
    availableNetworks: [ ChainIDs.Ethereum ],
    // this impossible value by defaut must be kept, it ensures that the `computed`
    // of `currentNetwork` selects the network of highest priority when `setCurrentNetwork()` has not been called yet
    currentNetwork: ChainIDs.Any,
  })
  return { data }
})

export function useNetworkStore() {
  const { data } = storeToRefs(store())

  const availableNetworks = computed(() => data.value.availableNetworks)
  const currentNetwork = computed(() => (useRuntimeConfig().public.chainIdByDefault ?? '0') as ChainId)
  const networkInfo = computed(() => ChainInfo[currentNetwork.value])

  function isNetworkDisabled(chainId: ChainId): boolean {
    // TODO: return `false` for everything once we are ready
    return (
      !useRuntimeConfig().public.showInDevelopment
      && chainId !== currentNetwork.value
    )
  }

  function setCurrentNetwork(chainId: ChainId) {
    data.value.currentNetwork = chainId
  }

  function isL1() {
    return ChainInfo[currentNetwork.value].L1 === currentNetwork.value
  }

  function epochsPerDay(): number {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined) {
      return 0
    }
    return (24 * 60 * 60) / (info.slotsPerEpoch * info.secondsPerSlot)
  }

  function epochToTs(epoch: number): number | undefined {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined || epoch < 0) {
      return undefined
    }
    return info.firstTimestamp + epoch * info.slotsPerEpoch * info.secondsPerSlot
  }

  const secondsPerEpoch = computed(() => {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined) {
      return -1
    }
    return info.slotsPerEpoch * info.secondsPerSlot
  })

  function slotToTs(slot: number) {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined || slot < 0) {
      return undefined
    }
    return info.firstTimestamp + slot * info.secondsPerSlot
  }

  function tsToSlot(timestampInSeconds: number) {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined) {
      return -1
    }
    return Math.floor((timestampInSeconds - info.firstTimestamp) / info.secondsPerSlot)
  }

  function slotToEpoch(slot: number) {
    const info = ChainInfo[currentNetwork.value]
    if (info.firstTimestamp === undefined) {
      return -1
    }
    return Math.floor(slot / info.slotsPerEpoch)
  }

  function tsToEpoch(ts: number) {
    return slotToEpoch(tsToSlot(ts))
  }

  return {
    availableNetworks,
    currentNetwork,
    epochsPerDay,
    epochToTs,
    isL1,
    isNetworkDisabled,
    networkInfo,
    secondsPerEpoch,
    setCurrentNetwork,
    slotToEpoch,
    slotToTs,
    tsToEpoch,
    tsToSlot,
  }
}
