import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import type { ApiDataResponse } from '~/types/api/common'
import { ChainIDs, ChainInfo, sortChainIDsByPriority } from '~/types/networks'

const { fetch } = useCustomFetch()

interface ApiChainInfo {
  chain_id: ChainIDs,
  name: string
}

const networkStore = defineStore('network-store', () => {
  const data = ref<{
    availableNetworks: ApiChainInfo[]
    currentNetwork: ChainIDs
  }>()
  return { data }
})

const { data: dataInternal } = storeToRefs(networkStore())

dataInternal.value!.availableNetworks = (await fetch<ApiDataResponse<ApiChainInfo[]>>(API_PATH.AVAILABLE_NETWORKS)).data
dataInternal.value!.currentNetwork = getAvailableNetworks()[0] // the network with the best priority in file `networks.ts` is chosen by default from the list

function getAvailableNetworks () : ChainIDs[] {
  return sortChainIDsByPriority(dataInternal.value!.availableNetworks.map(apiInfo => apiInfo.chain_id))
}

export function useNetworkStore () {
  const { data } = storeToRefs(networkStore())

  const currentNetwork = computed(() => data.value!.currentNetwork)
  const networkInfo = computed(() => ChainInfo[currentNetwork.value])

  function setCurrentNetwork (network: ChainIDs) {
    data.value!.currentNetwork = network
  }

  function epochsPerDay (): number {
    const info = ChainInfo[data.value!.currentNetwork]
    if (info.timeStampSlot0 === undefined) {
      return 0
    }
    return 24 * 60 * 60 / (info.slotsPerEpoch! * info.secondsPerSlot!)
  }

  function epochToTs (epoch: number): number | undefined {
    const info = ChainInfo[data.value!.currentNetwork]
    if (info.timeStampSlot0 === undefined || epoch < 0) {
      return undefined
    }

    return info.timeStampSlot0 + ((epoch * info.slotsPerEpoch!) * info.secondsPerSlot!)
  }

  function slotToTs (slot: number): number | undefined {
    const info = ChainInfo[data.value!.currentNetwork]
    if (info.timeStampSlot0 === undefined || slot < 0) {
      return undefined
    }

    return info.timeStampSlot0 + (slot * info.secondsPerSlot!)
  }

  function tsToSlot (ts: number): number {
    const info = ChainInfo[data.value!.currentNetwork]
    if (info.timeStampSlot0 === undefined) {
      return -1
    }
    return Math.floor((ts - info.timeStampSlot0) / info.secondsPerSlot!)
  }

  function slotToEpoch (slot: number): number {
    const info = ChainInfo[data.value!.currentNetwork]
    if (info.timeStampSlot0 === undefined) {
      return -1
    }
    return Math.floor(slot / info.slotsPerEpoch!)
  }

  return {
    getAvailableNetworks,
    currentNetwork,
    networkInfo,
    setCurrentNetwork,
    epochsPerDay,
    epochToTs,
    slotToTs,
    tsToSlot,
    slotToEpoch
  }
}
