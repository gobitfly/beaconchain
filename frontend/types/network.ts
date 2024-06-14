import type { CryptoCurrency } from '~/types/currencies'

export enum ChainFamily {
  Any = 'Any',
  Ethereum = 'Ethereum',
  Arbitrum = 'Arbitrum',
  Optimism = 'Optimism',
  Base = 'Base',
  Gnosis = 'Gnosis'
}

export enum ChainIDs {
  Any = 0, // to organize data internally (example of use: some ahead-results in the search bar belong to all networks)

  Ethereum = 1,
  Holesky = 17000,
  Sepolia = 11155111,

  ArbitrumOneEthereum = 42161,
  ArbitrumNovaEthereum= 42170,
  ArbitrumOneSepolia = 421614,

  OptimismEthereum = 10,
  OptimismSepolia = 11155420,

  BaseEthereum = 8453,
  BaseSepolia = 84532,

  Gnosis = 100,
  Chiado = 10200
}

interface ChainInfoFields {
  name: string,
  description: string,
  family: ChainFamily,
  mainNet: ChainIDs, // if the network is a testnet, this field points to the non-test network
  L1: ChainIDs, // if the network is a L2, this field points to the L1
  clCurrency: CryptoCurrency,
  elCurrency: CryptoCurrency,
  timeStampSlot0: number, // if this property is 0, it means that the network has no slots
  secondsPerSlot: number, // if this property is 0, it means that the network has no slots
  slotsPerEpoch: number, // if this property is 0, it means that the network has no slots
  path: string,
  priority: number // default order of the networks on the screen (ex: in the drop-down of the search bar)
}

export const ChainInfo: Record<ChainIDs, ChainInfoFields> = {
  [ChainIDs.Any]: {
    name: 'Any',
    description: 'Any network',
    family: ChainFamily.Any,
    mainNet: ChainIDs.Any,
    L1: ChainIDs.Any,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 12,
    slotsPerEpoch: 32,
    path: '/undefined',
    priority: 0 // data belonging to all networks is displayed first by default
  },

  [ChainIDs.Ethereum]: {
    name: 'Ethereum',
    description: 'Mainnet',
    family: ChainFamily.Ethereum,
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 1606824023,
    secondsPerSlot: 12,
    slotsPerEpoch: 32,
    path: '/ethereum',
    priority: 1
  },
  [ChainIDs.Holesky]: {
    name: 'Holesky',
    description: 'Testnet',
    family: ChainFamily.Ethereum,
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Holesky,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 1695902400,
    secondsPerSlot: 12,
    slotsPerEpoch: 32,
    path: '/holesky',
    priority: 2
  },
  [ChainIDs.Sepolia]: {
    name: 'Sepolia',
    description: 'Testnet',
    family: ChainFamily.Ethereum,
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 1655733600,
    secondsPerSlot: 12,
    slotsPerEpoch: 32,
    path: '/sepolia',
    priority: 3
  },

  [ChainIDs.ArbitrumOneEthereum]: {
    name: 'Arbitrum One',
    description: 'L2',
    family: ChainFamily.Arbitrum,
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/arbitrum-one-ethereum',
    priority: 10
  },
  [ChainIDs.ArbitrumNovaEthereum]: {
    name: 'Arbitrum Nova',
    description: 'L2',
    family: ChainFamily.Arbitrum,
    mainNet: ChainIDs.ArbitrumNovaEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/arbitrum-nova-ethereum',
    priority: 11
  },
  [ChainIDs.ArbitrumOneSepolia]: {
    name: 'Arbitrum Sepolia',
    description: 'Testnet',
    family: ChainFamily.Arbitrum,
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/arbitrum-one-sepolia',
    priority: 12
  },

  [ChainIDs.OptimismEthereum]: {
    name: 'Optimism',
    description: 'L2',
    family: ChainFamily.Optimism,
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/optimism-ethereum',
    priority: 20
  },
  [ChainIDs.OptimismSepolia]: {
    name: 'Optimism Sepolia',
    description: 'Testnet',
    family: ChainFamily.Optimism,
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/optimism-sepolia',
    priority: 21
  },

  [ChainIDs.BaseEthereum]: {
    name: 'Base',
    description: 'L2',
    family: ChainFamily.Base,
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/base-ethereum',
    priority: 30
  },
  [ChainIDs.BaseSepolia]: {
    name: 'Base Sepolia',
    description: 'Testnet',
    family: ChainFamily.Base,
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    timeStampSlot0: 0,
    secondsPerSlot: 0,
    slotsPerEpoch: 0,
    path: '/base-sepolia',
    priority: 31
  },

  [ChainIDs.Gnosis]: {
    name: 'Gnosis',
    description: 'Mainnet',
    family: ChainFamily.Gnosis,
    mainNet: ChainIDs.Gnosis,
    L1: ChainIDs.Gnosis,
    clCurrency: 'GNO',
    elCurrency: 'xDAI',
    timeStampSlot0: 1638993340,
    secondsPerSlot: 5,
    slotsPerEpoch: 16,
    path: '/gnosis',
    priority: 40
  },
  [ChainIDs.Chiado]: {
    name: 'Chiado',
    description: 'Testnet',
    family: ChainFamily.Gnosis,
    mainNet: ChainIDs.Gnosis,
    L1: ChainIDs.Chiado,
    clCurrency: 'GNO',
    elCurrency: 'xDAI',
    timeStampSlot0: 1665396300,
    secondsPerSlot: 5,
    slotsPerEpoch: 16,
    path: '/chiado',
    priority: 41
  }
}

export function getAllExistingChainIDs (sortByPriority : boolean) : ChainIDs[] {
  const list : ChainIDs[] = []

  for (const id in ChainIDs) {
    if (isNaN(Number(id))) {
      list.push(ChainIDs[id as keyof typeof ChainIDs])
    }
  }
  if (sortByPriority) {
    sortChainIDsByPriority(list)
  }
  return list
}

/**
 * Should be used only when you test a network different from the current one.
 * Whereever you would write `isMainNet(currentNetwork.value)` you should rather use `isMainNet()` from `useNetworkStore.ts`.
 */
export function isMainNet (network: ChainIDs) : boolean {
  return (ChainInfo[network].mainNet === network)
}

/**
 * Should be used only when you test a network different from the current one.
 * Wherever you would write `isL1(currentNetwork.value)` you should rather use `isL1()` from `useNetworkStore.ts`.
 */
export function isL1 (network: ChainIDs) : boolean {
  return (ChainInfo[network].L1 === network)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `epochsPerDay(currentNetwork.value)` you should rather use `epochsPerDay()` from `useNetworkStore.ts`.
 */
export function epochsPerDay (chainId: ChainIDs): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return 0
  }
  return 24 * 60 * 60 / (info.slotsPerEpoch * info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `epochToTs(currentNetwork.value, epoch)` you should rather use `epochToTs(epoch)` from `useNetworkStore.ts`.
 */
export function epochToTs (chainId: ChainIDs, epoch: number): number | undefined {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined || epoch < 0) {
    return undefined
  }

  return info.timeStampSlot0 + ((epoch * info.slotsPerEpoch) * info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `slotToTs(currentNetwork.value, slot)` you should rather use `slotToTs(slot)` from `useNetworkStore.ts`.
 */
export function slotToTs (chainId: ChainIDs, slot: number): number | undefined {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined || slot < 0) {
    return undefined
  }

  return info.timeStampSlot0 + (slot * info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `tsToSlot(currentNetwork.value, ts)` you should rather use `tsToSlot(ts)` from `useNetworkStore.ts`.
 */
export function tsToSlot (chainId: ChainIDs, ts: number): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return -1
  }
  return Math.floor((ts - info.timeStampSlot0) / info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `slotToEpoch(currentNetwork.value, slot)` you should rather use `slotToEpoch(slot)` from `useNetworkStore.ts`.
 */
export function slotToEpoch (chainId: ChainIDs, slot: number): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return -1
  }
  return Math.floor(slot / info.slotsPerEpoch)
}

/**
 * @param list List to sort. Its order will be modified because the function sorts in place.
 * @returns List sorted in place, so the same as parameter `list`.
 */
export function sortChainIDsByPriority (list : ChainIDs[]) : ChainIDs[] {
  return list.sort((a, b) => ChainInfo[a].priority - ChainInfo[b].priority)
}
