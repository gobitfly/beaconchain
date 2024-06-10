import { API_PATH } from '~/types/customFetch'
import type { ApiDataResponse } from '~/types/api/common'
import * as networkTs from '~/types/network'

interface ApiChainInfo {
  chain_id: networkTs.ChainIDs,
  name: string
}

// in the parentheses are temporary values so the rest of the front-end doesn't crash until these variables are filled with actual data from the the API response
const networkList = ref<ApiChainInfo[]>([{ chain_id: networkTs.ChainIDs.Ethereum, name: '' }])
const networkChoice = ref<networkTs.ChainIDs>(networkTs.ChainIDs.Ethereum)
let currentNetworkHasBeenChosen = false

function fillDataFromResponse (response : ApiDataResponse<ApiChainInfo[]>) {
  networkList.value = response.data.sort((a, b) => networkTs.ChainInfo[a.chain_id].priority - networkTs.ChainInfo[b.chain_id].priority)
  if (!currentNetworkHasBeenChosen) {
    networkChoice.value = networkList.value[0].chain_id // the network with the best priority in file `networks.ts` is chosen by default from the list
  }
}

function setCurrentNetwork (chainId: networkTs.ChainIDs) {
  networkChoice.value = chainId
  currentNetworkHasBeenChosen = true
}

export function useNetwork () {
  if (!networkList.value[0].name) {
    const { fetch } = useCustomFetch()
    fetch<ApiDataResponse<ApiChainInfo[]>>(API_PATH.AVAILABLE_NETWORKS).then(fillDataFromResponse)
  }

  const availableNetworks = computed(() => networkList.value.map(apiInfo => apiInfo.chain_id))
  const currentNetwork = computed(() => networkChoice.value)
  const networkInfo = computed(() => networkTs.ChainInfo[networkChoice.value])

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
