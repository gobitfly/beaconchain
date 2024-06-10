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
    availableNetworks: ApiChainInfo[],
    currentNetwork: networkTs.ChainIDs,
    availableNetworksHasBeenFilled: boolean,
    currentNetworkHasBeenChosen: boolean
  }>({
    // The values below are temporary and get replaced with actual data when the API responds.
    // In the meantime, they allow the front-end to run for a few seconds with default values.
    availableNetworks: [{ chain_id: networkTs.ChainIDs.Ethereum, name: 'Ethereum' }],
    currentNetwork: networkTs.ChainIDs.Ethereum,
    availableNetworksHasBeenFilled: false,
    currentNetworkHasBeenChosen: false
  })
  return { data }
})

export function useNetworkStore () {
  const { data } = storeToRefs(store())

  if (!data.value.availableNetworksHasBeenFilled) {
    const { fetch } = useCustomFetch()
    data.value.availableNetworksHasBeenFilled = true // do not put it inside the `then()` otherwise the calls to `useNetworkStore()` made before the API responds will call the API multiple times
    fetch<ApiDataResponse<ApiChainInfo[]>>(API_PATH.AVAILABLE_NETWORKS).then((response) => {
      data.value.availableNetworks = response.data.sort((a, b) => networkTs.ChainInfo[a.chain_id].priority - networkTs.ChainInfo[b.chain_id].priority)
      if (!data.value.currentNetworkHasBeenChosen) {
        // by default, the current network is the one with the best priority in file `networks.ts`
        data.value.currentNetwork = data.value.availableNetworks[0].chain_id
      }
    })
  }

  const availableNetworks = computed(() => data.value.availableNetworks.map(apiInfo => apiInfo.chain_id))
  const currentNetwork = computed(() => data.value.currentNetwork)
  const networkInfo = computed(() => networkTs.ChainInfo[data.value.currentNetwork])

  function setCurrentNetwork (chainId: networkTs.ChainIDs) {
    data.value.currentNetwork = chainId
    data.value.currentNetworkHasBeenChosen = true
  }

  function isMainNet () : boolean {
    return networkTs.isMainNet(currentNetwork.value)
  }

  function isL1 () : boolean {
    return networkTs.isL1(currentNetwork.value)
  }

  function epochsPerDay (): number {
    return networkTs.epochsPerDay(currentNetwork.value)
  }

  function epochToTs (epoch: number): number | undefined {
    return networkTs.epochToTs(currentNetwork.value, epoch)
  }

  function slotToTs (slot: number): number | undefined {
    return networkTs.slotToTs(currentNetwork.value, slot)
  }

  function tsToSlot (ts: number): number {
    return networkTs.tsToSlot(currentNetwork.value, ts)
  }

  function slotToEpoch (slot: number): number {
    return networkTs.slotToEpoch(currentNetwork.value, slot)
  }

  return {
    availableNetworks,
    currentNetwork,
    networkInfo,
    setCurrentNetwork,
    isMainNet,
    isL1,
    epochsPerDay,
    epochToTs,
    slotToTs,
    tsToSlot,
    slotToEpoch
  }
}
