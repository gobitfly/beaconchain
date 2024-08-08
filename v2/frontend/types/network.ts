import type { CryptoCurrency } from '~/types/currencies'

export enum ChainFamily {
  Any = 'Any',
  Arbitrum = 'Arbitrum',
  Base = 'Base',
  Ethereum = 'Ethereum',
  Gnosis = 'Gnosis',
  Optimism = 'Optimism',
}

export enum ChainIDs {
  Any = 0, // to organize data internally (example of use: some ahead-results in the search bar belong to all networks)

  ArbitrumNovaEthereum = 42170,
  ArbitrumOneEthereum = 42161,
  ArbitrumOneSepolia = 421614,

  BaseEthereum = 8453,
  BaseSepolia = 84532,
  Chiado = 10200,

  Ethereum = 1,
  Gnosis = 100,

  Holesky = 17000,
  OptimismEthereum = 10,

  OptimismSepolia = 11155420,
  Sepolia = 11155111,
}

export interface ChainInfoFields {
  clCurrency: CryptoCurrency
  description: string
  elCurrency: CryptoCurrency
  family: ChainFamily
  L1: ChainIDs // if the network is a L2, this field points to the L1
  mainNet: ChainIDs // if the network is a testnet, this field points to the non-test network
  name: string
  nameParts: string[]
  priority: number // default order of the networks on the screen (ex: in the drop-down of the search bar)
  secondsPerSlot: number // if this property is 0, it means that the network has no slots
  shortName: string
  slotsPerEpoch: number // if this property is 0, it means that the network has no slots
  timeStampSlot0: number // if this property is 0, it means that the network has no slots
}

export const ChainInfo: Record<ChainIDs, ChainInfoFields> = {
  [ChainIDs.Any]: {
    clCurrency: 'ETH',
    description: 'Any network',
    elCurrency: 'ETH',
    family: ChainFamily.Any,
    L1: ChainIDs.Any,
    mainNet: ChainIDs.Any,
    name: 'Any network',
    nameParts: ['Any', 'network'],
    priority: 0, // data belonging to all networks is displayed first by default
    secondsPerSlot: 12,
    shortName: 'Any',
    slotsPerEpoch: 32,
    timeStampSlot0: 0,
  },

  [ChainIDs.ArbitrumNovaEthereum]: {
    clCurrency: 'ETH',
    description: 'L2',
    elCurrency: 'ETH',
    family: ChainFamily.Arbitrum,
    L1: ChainIDs.Ethereum,
    mainNet: ChainIDs.ArbitrumNovaEthereum,
    name: 'Arbitrum Nova',
    nameParts: ['Arbitrum Nova', ''],
    priority: 11,
    secondsPerSlot: 0,
    shortName: 'Arbitrum',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },
  [ChainIDs.ArbitrumOneEthereum]: {
    clCurrency: 'ETH',
    description: 'L2',
    elCurrency: 'ETH',
    family: ChainFamily.Arbitrum,
    L1: ChainIDs.Ethereum,
    mainNet: ChainIDs.ArbitrumOneEthereum,
    name: 'Arbitrum One',
    nameParts: ['Arbitrum One', ''],
    priority: 10,
    secondsPerSlot: 0,
    shortName: 'Arbitrum',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },
  [ChainIDs.ArbitrumOneSepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Arbitrum,
    L1: ChainIDs.Sepolia,
    mainNet: ChainIDs.ArbitrumOneEthereum,
    name: 'Arbitrum Sepolia',
    nameParts: ['Arbitrum', 'Sepolia'],
    priority: 12,
    secondsPerSlot: 0,
    shortName: 'Arbitrum',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },

  [ChainIDs.BaseEthereum]: {
    clCurrency: 'ETH',
    description: 'L2',
    elCurrency: 'ETH',
    family: ChainFamily.Base,
    L1: ChainIDs.Ethereum,
    mainNet: ChainIDs.BaseEthereum,
    name: 'Base',
    nameParts: ['Base', ''],
    priority: 30,
    secondsPerSlot: 0,
    shortName: 'Base',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },
  [ChainIDs.BaseSepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Base,
    L1: ChainIDs.Sepolia,
    mainNet: ChainIDs.BaseEthereum,
    name: 'Base Sepolia',
    nameParts: ['Base', 'Sepolia'],
    priority: 31,
    secondsPerSlot: 0,
    shortName: 'Base',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },
  [ChainIDs.Chiado]: {
    clCurrency: 'GNO',
    description: 'Testnet',
    elCurrency: 'xDAI',
    family: ChainFamily.Gnosis,
    L1: ChainIDs.Chiado,
    mainNet: ChainIDs.Gnosis,
    name: 'Gnosis Chiado',
    nameParts: ['Gnosis', 'Chiado'],
    priority: 41,
    secondsPerSlot: 5,
    shortName: 'Chiado',
    slotsPerEpoch: 16,
    timeStampSlot0: 1665396300,
  },

  [ChainIDs.Ethereum]: {
    clCurrency: 'ETH',
    description: 'Mainnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    L1: ChainIDs.Ethereum,
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum',
    nameParts: ['Ethereum', ''],
    priority: 1,
    secondsPerSlot: 12,
    shortName: 'Ethereum',
    slotsPerEpoch: 32,
    timeStampSlot0: 1606824023,
  },
  [ChainIDs.Gnosis]: {
    clCurrency: 'GNO',
    description: 'Mainnet',
    elCurrency: 'xDAI',
    family: ChainFamily.Gnosis,
    L1: ChainIDs.Gnosis,
    mainNet: ChainIDs.Gnosis,
    name: 'Gnosis',
    nameParts: ['Gnosis', ''],
    priority: 40,
    secondsPerSlot: 5,
    shortName: 'Gnosis',
    slotsPerEpoch: 16,
    timeStampSlot0: 1638993340,
  },

  [ChainIDs.Holesky]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    L1: ChainIDs.Holesky,
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum Holesky',
    nameParts: ['Ethereum', 'Holesky'],
    priority: 2,
    secondsPerSlot: 12,
    shortName: 'Holesky',
    slotsPerEpoch: 32,
    timeStampSlot0: 1695902400,
  },
  [ChainIDs.OptimismEthereum]: {
    clCurrency: 'ETH',
    description: 'L2',
    elCurrency: 'ETH',
    family: ChainFamily.Optimism,
    L1: ChainIDs.Ethereum,
    mainNet: ChainIDs.OptimismEthereum,
    name: 'Optimism',
    nameParts: ['Optimism', ''],
    priority: 20,
    secondsPerSlot: 0,
    shortName: 'Optimism',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },

  [ChainIDs.OptimismSepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Optimism,
    L1: ChainIDs.Sepolia,
    mainNet: ChainIDs.OptimismEthereum,
    name: 'Optimism Sepolia',
    nameParts: ['Optimism', 'Sepolia'],
    priority: 21,
    secondsPerSlot: 0,
    shortName: 'Optimism',
    slotsPerEpoch: 0,
    timeStampSlot0: 0,
  },
  [ChainIDs.Sepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    L1: ChainIDs.Sepolia,
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum Sepolia',
    nameParts: ['Ethereum', 'Sepolia'],
    priority: 3,
    secondsPerSlot: 12,
    shortName: 'Sepolia',
    slotsPerEpoch: 32,
    timeStampSlot0: 1655733600,
  },
}

export function getAllExistingChainIDs(sortByPriority: boolean): ChainIDs[] {
  const list: ChainIDs[] = []

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
 * Whereever you would write `isMainNet(currentNetwork.value)` you should
 * rather use `isMainNet()` from `useNetworkStore.ts`.
 */
export function isMainNet(network: ChainIDs): boolean {
  return ChainInfo[network].mainNet === network
}

/**
 * Should be used only when you test a network different from the current one.
 * Wherever you would write `isL1(currentNetwork.value)` you should rather use `isL1()` from `useNetworkStore.ts`.
 */
export function isL1(network: ChainIDs): boolean {
  return ChainInfo[network].L1 === network
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `epochsPerDay(currentNetwork.value)` you should
 * rather use `epochsPerDay()` from `useNetworkStore.ts`.
 */
export function epochsPerDay(chainId: ChainIDs): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return 0
  }
  return (24 * 60 * 60) / (info.slotsPerEpoch * info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `epochToTs(currentNetwork.value, epoch)` you should
 *  rather use `epochToTs(epoch)` from `useNetworkStore.ts`.
 */
export function epochToTs(
  chainId: ChainIDs,
  epoch: number,
): number | undefined {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined || epoch < 0) {
    return undefined
  }

  return info.timeStampSlot0 + epoch * info.slotsPerEpoch * info.secondsPerSlot
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `slotToTs(currentNetwork.value, slot)` you should
 *  rather use `slotToTs(slot)` from `useNetworkStore.ts`.
 */
export function slotToTs(chainId: ChainIDs, slot: number): number | undefined {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined || slot < 0) {
    return undefined
  }

  return info.timeStampSlot0 + slot * info.secondsPerSlot
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `tsToSlot(currentNetwork.value, ts)` you should
 * rather use `tsToSlot(ts)` from `useNetworkStore.ts`.
 */
export function tsToSlot(chainId: ChainIDs, ts: number): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return -1
  }
  return Math.floor((ts - info.timeStampSlot0) / info.secondsPerSlot)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `slotToEpoch(currentNetwork.value, slot)` you should
 *  rather use `slotToEpoch(slot)` from `useNetworkStore.ts`.
 */
export function slotToEpoch(chainId: ChainIDs, slot: number): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return -1
  }
  return Math.floor(slot / info.slotsPerEpoch)
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `secondsPerEpoch(currentNetwork.value)` you should
 * rather use `secondsPerEpoch()` from `useNetworkStore.ts`.
 */
export function secondsPerEpoch(chainId: ChainIDs): number {
  const info = ChainInfo[chainId]
  if (info.timeStampSlot0 === undefined) {
    return -1
  }
  return info.slotsPerEpoch * info.secondsPerSlot
}

/**
 * @param list List to sort. Its order will be modified because the function sorts in place.
 * @returns List sorted in place, so the same as parameter `list`.
 */
export function sortChainIDsByPriority(list: ChainIDs[]): ChainIDs[] {
  return list.sort((a, b) => ChainInfo[a].priority - ChainInfo[b].priority)
}
