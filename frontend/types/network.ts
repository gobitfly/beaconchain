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
  Pectra_Devnet_5: 7088110746,
  Sepolia: 11155111,
} as const

export type ChainId = (typeof ChainIDs)[keyof typeof ChainIDs]

export interface ChainInfoFields {
  clCurrency: CurrencyCodeCrypto,
  description: string,
  displayCurrencyDefault: DisplayCurrency,
  elCurrency: CurrencyCodeCrypto,
  family: ChainFamily,
  hasRocketPool: boolean,
  mainCurrency: CurrencyCodeCrypto,
  mainNet: ChainId,
  name: string,
  nameParts: string[],
  priority: number, // default order of the networks on the screen (ex: in the drop-down of the search bar)
  secondsPerSlot: number, // if this property is 0, it means that the network has no slots
  shortName: string,
  slotsPerEpoch: number, // if this property is 0, it means that the network has no slots
  timeStampSlot0: number, // if this property is 0, it means that the network has no slots
}

export type DisplayCurrency = {
  consensusLayer: 'ETH',
  executionLayer: 'ETH',
  fiat: 'USD',
  main: 'ETH',
} | {
  consensusLayer: 'GNO',
  executionLayer: 'xDAI',
  fiat: 'USD',
  main: 'GNO',
}

export const ChainInfo: Record<ChainId, ChainInfoFields> = {
  [ChainIDs.Any]: {
    clCurrency: 'ETH',
    description: 'Any network',
    displayCurrencyDefault: {
      consensusLayer: 'ETH',
      executionLayer: 'ETH',
      fiat: 'USD',
      main: 'ETH',
    },
    elCurrency: 'ETH',
    family: ChainFamily.Any,
    hasRocketPool: false,
    mainCurrency: 'ETH',
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
    displayCurrencyDefault: {
      consensusLayer: 'ETH',
      executionLayer: 'ETH',
      fiat: 'USD',
      main: 'ETH',
    },
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    hasRocketPool: true,
    mainCurrency: 'ETH',
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
    clCurrency: 'mGNO',
    description: '',
    displayCurrencyDefault: {
      consensusLayer: 'GNO',
      executionLayer: 'xDAI',
      fiat: 'USD',
      main: 'GNO',
    },
    elCurrency: 'xDAI',
    family: ChainFamily.Gnosis,
    hasRocketPool: false,
    mainCurrency: 'GNO',
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
    displayCurrencyDefault: {
      consensusLayer: 'ETH',
      executionLayer: 'ETH',
      fiat: 'USD',
      main: 'ETH',
    },
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    hasRocketPool: true,
    mainCurrency: 'ETH',
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
  [ChainIDs.Pectra_Devnet_5]: {
    clCurrency: 'ETH',
    description: 'Devnet',
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    mainCurrency: 'ETH',
    mainNet: ChainIDs.Ethereum,
    name: 'Ethereum Pectra Devnet 5',
    nameParts: [
      'Ethereum',
      'Pectra',
      'Devnet',
      '5',
    ],
    priority: 41,
    secondsPerSlot: 12,
    shortName: 'Pectra',
    slotsPerEpoch: 32,
    timeStampSlot0: 1737034260,
  },
  [ChainIDs.Sepolia]: {
    clCurrency: 'ETH',
    description: 'Testnet',
    displayCurrencyDefault: {
      consensusLayer: 'ETH',
      executionLayer: 'ETH',
      fiat: 'USD',
      main: 'ETH',
    },
    elCurrency: 'ETH',
    family: ChainFamily.Ethereum,
    hasRocketPool: false,
    mainCurrency: 'ETH',
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
