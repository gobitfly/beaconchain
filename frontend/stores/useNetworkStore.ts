import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import type { ApiDataResponse } from '~/types/api/common'
import * as networkTs from '~/types/network'

interface ApiChainInfo {
  chain_id: networkTs.ChainIDs,
  name: string
}

const store = defineStore('network-store', () => {
  const data = ref<{
    availableNetworks: networkTs.ChainIDs[],
    currentNetwork: networkTs.ChainIDs
  }>({
    availableNetworks: [networkTs.ChainIDs.Ethereum],
    currentNetwork: networkTs.ChainIDs.Any // this impossible value by defaut must be kept, it ensures that the `computed` of `currentNetwork` selects the network of highest priority when `setCurrentNetwork()` has not been called yet
  })
  return { data }
})

export function useNetworkStore () {
  const { data } = storeToRefs(store())

  /**
   * Needs to be called once, when the front-end is loading. Unnecessary afterwards.
   */
  async function loadAvailableNetworks () : Promise<boolean> {
    try {
      const { fetch } = useCustomFetch()
      const response = await fetch<ApiDataResponse<ApiChainInfo[]>>(API_PATH.AVAILABLE_NETWORKS)
      if (!response.data || !response.data.length) {
        return false
      }
      data.value.availableNetworks = networkTs.sortChainIDsByPriority(response.data.map(apiInfo => apiInfo.chain_id))
      return true
    } catch {
      return false
    }
  }

  const availableNetworks = computed(() => data.value.availableNetworks)
  const currentNetwork = computed(() => availableNetworks.value.includes(data.value.currentNetwork) ? data.value.currentNetwork : availableNetworks.value[0])
  const networkInfo = computed(() => networkTs.ChainInfo[currentNetwork.value])

  function isNetworkDisabled (chainId: networkTs.ChainIDs) : boolean {
    // TODO: return `false` for everything once we are ready
    return !useRuntimeConfig().public.showInDevelopment && chainId !== currentNetwork.value
  }

  function setCurrentNetwork (chainId: networkTs.ChainIDs) {
    data.value.currentNetwork = chainId
  }

  function isMainNet () : boolean {
    return networkTs.isMainNet(currentNetwork.value)
  }

  function isL1 () : boolean {
    return networkTs.isL1(currentNetwork.value)
  }

  function epochsPerDay () : number {
    return networkTs.epochsPerDay(currentNetwork.value)
  }

  function epochToTs (epoch: number) : number | undefined {
    return networkTs.epochToTs(currentNetwork.value, epoch)
  }

  function secondsPerEpoch ():number {
    return networkTs.secondsPerEpoch(currentNetwork.value)
  }

  function slotToTs (slot: number) : number | undefined {
    return networkTs.slotToTs(currentNetwork.value, slot)
  }

  function tsToSlot (ts: number) : number {
    return networkTs.tsToSlot(currentNetwork.value, ts)
  }

  function slotToEpoch (slot: number) : number {
    return networkTs.slotToEpoch(currentNetwork.value, slot)
  }

  function tsToEpoch (ts: number): number {
    return slotToEpoch(tsToSlot(ts))
  }

  return {
    loadAvailableNetworks,
    availableNetworks,
    isNetworkDisabled,
    currentNetwork,
    networkInfo,
    setCurrentNetwork,
    isMainNet,
    isL1,
    epochsPerDay,
    epochToTs,
    secondsPerEpoch,
    slotToTs,
    tsToSlot,
    tsToEpoch,
    slotToEpoch
  }
}
