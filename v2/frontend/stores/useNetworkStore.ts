import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import { ChainIDs, ChainInfo } from '~/types/networks'

const { fetch } = useCustomFetch()

interface ApiChainInfo {
  chain_id: ChainIDs,
  name: string
}

const networkStore = defineStore('network-store', () => {
  const data = ref<{
    available: ApiChainInfo[]
    current: ChainIDs
  } | undefined | null>()
  return { data }
})

networkStore().data!.available = await fetch<ApiChainInfo[]>(API_PATH.AVAILABLE_NETWORKS)

export function useNetworkStore () {
  const { data } = storeToRefs(networkStore())

  function epochToTs (epoch: number): number | undefined {
    if (epoch < 0) {
      return undefined
    }

    return tsForSlot0 + ((epoch * slotsPerEpoch) * secondsPerSlot)
  }

  function slotToTs (slot: number): number | undefined {
    if (slot < 0) {
      return undefined
    }

    return tsForSlot0 + (slot * secondsPerSlot)
  }

  function epochsPerDay (): number {
    return 24 * 60 * 60 / (slotsPerEpoch * secondsPerSlot)
  }

  function slotToEpoch (slot: number): number {
    return Math.floor(slot / slotsPerEpoch)
  }

  return { epochToTs, epochsPerDay, slotsPerEpoch, slotToTs, slotToEpoch }
}

export function isMainNet (network: ChainIDs) : boolean {
  return (ChainInfo[network].mainNet === network)
}

export function isL1 (network: ChainIDs) : boolean {
  return (ChainInfo[network].L1 === network)
}

export function sortChainIDsByPriority (list : ChainIDs[]) {
  list.sort((a, b) => { return ChainInfo[a].priority - ChainInfo[b].priority })
}

// TODO: request it from the API
export function getListOfImplementedChainIDs (sortByPriority : boolean) : ChainIDs[] {
  const list = [ChainIDs.Ethereum, ChainIDs.ArbitrumOneEthereum, ChainIDs.OptimismEthereum, ChainIDs.BaseEthereum, ChainIDs.Gnosis]
  if (sortByPriority) {
    sortChainIDsByPriority(list)
  }
  return list
}
