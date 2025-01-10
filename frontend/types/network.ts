import type { CryptoCurrency } from '~/types/currencies'

export enum ChainFamily {
  Any = 'Any',
  Arbitrum = 'Arbitrum',
  Base = 'Base',
  Ethereum = 'Ethereum',
  Gnosis = 'Gnosis',
  Optimism = 'Optimism',
}

const ChainIDs = {
  Any: 0, // to organize data internally (example of use: some ahead-results in the search bar belong to all networks)

  Ethereum: 1,
  Gnosis: 100,

  Holesky: 17000,

  Sepolia: 11155111,
} as const

export type ChainId = (typeof ChainIDs)[keyof typeof ChainIDs]

export interface ChainInfoFields {
  clCurrency: CryptoCurrency,
  description: string,
  elCurrency: CryptoCurrency,
  family: ChainFamily,
  mainNet: ChainId,
  name: string,
  nameParts: string[],
  priority: number, // default order of the networks on the screen (ex: in the drop-down of the search bar)
  secondsPerSlot: number, // if this property is 0, it means that the network has no slots
  shortName: string,
  slotsPerEpoch: number, // if this property is 0, it means that the network has no slots
  timeStampSlot0: number, // if this property is 0, it means that the network has no slots
}

export const ChainInfo: Record<ChainId, ChainInfoFields> = {
  [ChainIDs.Any]: {
    clCurrency: 'ETH',
    description: 'Any network',
    elCurrency: 'ETH',
    family: ChainFamily.Any,
    mainNet: ChainIDs.Any,
    name: 'Any network',
    nameParts: [
      'Any',
      'network',
    ],
    priority: 0, // data belonging to all networks is displayed first by default
    secondsPerSlot: 12,
    shortName: 'Any',
    slotsPerEpoch: 32,
    timeStampSlot0: 0,
  },
  [ChainIDs.Ethereum]: {
    clCurrency: 'ETH',
    description: 'Mainnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum',
    nameParts: [
      'Ethereum',
      '',
    ],
    priority: 1,
    secondsPerSlot: 12,
    shortName: 'Ethereum',
    slotsPerEpoch: 32,
    timeStampSlot0: 1606824023,
  },
  [ChainIDs.Gnosis]: {
    clCurrency: 'GNO',
    description: '',
    elCurrency: 'xDAI',
    family: ChainFamily.Gnosis,
    mainNet: ChainIDs.Gnosis,
    name: 'Gnosis',
    nameParts: [
      'Gnosis',
      '',
    ],
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
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum Holesky',
    nameParts: [
      'Ethereum',
      'Holesky',
    ],
    priority: 2,
    secondsPerSlot: 12,
    shortName: 'Holesky',
    slotsPerEpoch: 32,
    timeStampSlot0: 1695902400,
  },
  [ChainIDs.Sepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum Sepolia',
    nameParts: [
      'Ethereum',
      'Sepolia',
    ],
    priority: 3,
    secondsPerSlot: 12,
    shortName: 'Sepolia',
    slotsPerEpoch: 32,
    timeStampSlot0: 1655733600,
  },
} as const
