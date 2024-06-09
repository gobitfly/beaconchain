import { API_PATH, mapping } from '~/types/customFetch'
import type { ApiDataResponse } from '~/types/api/common'
import { ChainIDs, ChainInfo } from '~/types/networks'

interface ApiChainInfo {
  chain_id: ChainIDs,
  name: string
}

// in the parentheses are temporary values so the rest of the front-end doesn't crash until these variables are filled with actual data from the the API response
const networkList = ref<ApiChainInfo[]>([{ chain_id: ChainIDs.Ethereum, name: '' }])
const networkChoice = ref<ChainIDs>(ChainIDs.Ethereum)
let currentNetworkHasBeenChosen = false

function fillDataFromResponse (response : ApiDataResponse<ApiChainInfo[]>) {
  networkList.value = response.data.sort((a, b) => ChainInfo[a.chain_id].priority - ChainInfo[b.chain_id].priority)
  if (!currentNetworkHasBeenChosen) {
    networkChoice.value = networkList.value[0].chain_id // the network with the best priority in file `networks.ts` is chosen by default from the list
  }
}

// The following block calls the API or mock its response.
// `customFetch` could not work here because it needs Nuxt features, which are not enabled when `useNetworks` is called from a composable for example
const dataAccess = mapping[API_PATH.AVAILABLE_NETWORKS]
if (dataAccess.mock) {
  fillDataFromResponse(dataAccess.mockFunction!())
} else {
  fetch(dataAccess.path).then((resp) => {
    resp.json().then(fillDataFromResponse)
  })
}

function setCurrentNetwork (chainId: ChainIDs) {
  networkChoice.value = chainId
  currentNetworkHasBeenChosen = true
}

export function useNetwork () {
  const availableNetworks = computed(() => networkList.value.map(apiInfo => apiInfo.chain_id))
  const currentNetwork = computed(() => networkChoice.value)
  const networkInfo = computed(() => ChainInfo[networkChoice.value])

  function epochsPerDay (): number {
    const info = networkInfo.value
    if (info.timeStampSlot0 === undefined) {
      return 0
    }
    return 24 * 60 * 60 / (info.slotsPerEpoch * info.secondsPerSlot)
  }

  function epochToTs (epoch: number): number | undefined {
    const info = networkInfo.value
    if (info.timeStampSlot0 === undefined || epoch < 0) {
      return undefined
    }

    return info.timeStampSlot0 + ((epoch * info.slotsPerEpoch) * info.secondsPerSlot)
  }

  function slotToTs (slot: number): number | undefined {
    const info = networkInfo.value
    if (info.timeStampSlot0 === undefined || slot < 0) {
      return undefined
    }

    return info.timeStampSlot0 + (slot * info.secondsPerSlot)
  }

  function tsToSlot (ts: number): number {
    const info = networkInfo.value
    if (info.timeStampSlot0 === undefined) {
      return -1
    }
    return Math.floor((ts - info.timeStampSlot0) / info.secondsPerSlot)
  }

  function slotToEpoch (slot: number): number {
    const info = networkInfo.value
    if (info.timeStampSlot0 === undefined) {
      return -1
    }
    return Math.floor(slot / info.slotsPerEpoch)
  }

  return {
    availableNetworks,
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
