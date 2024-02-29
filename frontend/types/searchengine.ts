import { ChainIDs } from '~/types/networks'

export type SearchBarStyle = 'discreet' | 'gaudy'

export enum Categories {
  Tokens = 'tokens',
  NFTs = 'nfts',
  Protocol = 'protocol',
  Addresses = 'addresses',
  Validators = 'validators'
}

export enum ResultTypes {
  Tokens = 'tokens',
  NFTs = 'nfts',
  Epochs = 'epochs',
  Slots = 'slots',
  Blocks = 'blocks',
  BlockRoots = 'block_roots',
  StateRoots = 'state_roots',
  Transactions = 'transactions',
  TransactionBatches = 'transaction_batches',
  StateBatches = 'state_batches',
  Accounts = 'accounts',
  Contracts = 'contracts',
  Ens = 'ens_names',
  EnsOverview = 'ens_overview',
  Graffiti = 'graffiti',
  ValidatorsByIndex = 'validators_by_index',
  ValidatorsByPubkey = 'validators_by_pubkey',
  ValidatorsByDepositAddress = 'validators_by_deposit_address',
  ValidatorsByDepositEnsName = 'validators_by_deposit_ens_name',
  ValidatorsByWithdrawalCredential = 'validators_by_withdrawal_credential',
  ValidatorsByWithdrawalAddress = 'validators_by_withdrawal_address',
  ValidatorsByWithdrawalEnsName = 'validators_by_withdrawal_ens_name',
  ValidatorsByGraffiti = 'validators_by_graffiti',
  ValidatorsByName = 'validators_by_name'
}

interface CategoryInfoFields {
  filterLabel : string
}

export const CategoryInfo: Record<Categories, CategoryInfoFields> = {
  [Categories.Tokens]: { filterLabel: 'Tokens' },
  [Categories.NFTs]: { filterLabel: 'NFTs' },
  [Categories.Protocol]: { filterLabel: 'Protocol' },
  [Categories.Addresses]: { filterLabel: 'Addresses' },
  [Categories.Validators]: { filterLabel: 'Validators' }
}

interface TypeInfoFields {
  title: string,
  titleShort: string,
  logo: string,
  category: Categories,
  priority: number,
  belongsToAllNetworks: boolean
}

export const TypeInfo: Record<ResultTypes, TypeInfoFields> = {
  [ResultTypes.Tokens]: {
    title: 'ERC-20 token',
    titleShort: 'Token',
    logo: '',
    category: Categories.Tokens,
    priority: 3,
    belongsToAllNetworks: true
  },
  [ResultTypes.NFTs]: {
    title: 'NFT (ERC-721 & ERC-1155 token)',
    titleShort: 'NFT',
    logo: '',
    category: Categories.NFTs,
    priority: 4,
    belongsToAllNetworks: true
  },
  [ResultTypes.Epochs]: {
    title: 'Epoch',
    titleShort: 'Epoch',
    logo: '',
    category: Categories.Protocol,
    priority: 12,
    belongsToAllNetworks: false
  },
  [ResultTypes.Slots]: {
    title: 'Slot',
    titleShort: 'Slot',
    logo: '',
    category: Categories.Protocol,
    priority: 11,
    belongsToAllNetworks: false
  },
  [ResultTypes.Blocks]: {
    title: 'Block',
    titleShort: 'Block',
    logo: '',
    category: Categories.Protocol,
    priority: 10,
    belongsToAllNetworks: false
  },
  [ResultTypes.BlockRoots]: {
    title: 'Block root',
    titleShort: 'Block Root',
    logo: '',
    category: Categories.Protocol,
    priority: 18,
    belongsToAllNetworks: false
  },
  [ResultTypes.StateRoots]: {
    title: 'State root',
    titleShort: 'State Root',
    logo: '',
    category: Categories.Protocol,
    priority: 19,
    belongsToAllNetworks: false
  },
  [ResultTypes.Transactions]: {
    title: 'Transaction',
    titleShort: 'Transaction',
    logo: '',
    category: Categories.Protocol,
    priority: 17,
    belongsToAllNetworks: false
  },
  [ResultTypes.TransactionBatches]: {
    title: 'Transaction batch',
    titleShort: 'TX Batch',
    logo: '',
    category: Categories.Protocol,
    priority: 14,
    belongsToAllNetworks: false
  },
  [ResultTypes.StateBatches]: {
    title: 'State batch',
    titleShort: 'State Batch',
    logo: '',
    category: Categories.Protocol,
    priority: 13,
    belongsToAllNetworks: false
  },
  [ResultTypes.Accounts]: {
    title: 'Account',
    titleShort: 'Account',
    logo: '',
    category: Categories.Addresses,
    priority: 2,
    belongsToAllNetworks: true
  },
  [ResultTypes.Contracts]: {
    title: 'Contract',
    titleShort: 'Contract',
    logo: '',
    category: Categories.Addresses,
    priority: 2,
    belongsToAllNetworks: true
  },
  [ResultTypes.Ens]: {
    title: 'ENS address',
    titleShort: 'ENS',
    logo: '',
    category: Categories.Addresses,
    priority: 1,
    belongsToAllNetworks: true
  },
  [ResultTypes.EnsOverview]: {
    title: 'Overview of ENS domain',
    titleShort: 'ENS Overview',
    logo: '',
    category: Categories.Addresses,
    priority: 15,
    belongsToAllNetworks: true
  },
  [ResultTypes.Graffiti]: {
    title: 'Graffito',
    titleShort: 'Graffito',
    logo: '',
    category: Categories.Protocol,
    priority: 16,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByIndex]: {
    title: 'Validator by index',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByPubkey]: {
    title: 'Validator by public key',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByDepositAddress]: {
    title: 'Validator by deposit address',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 6,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByDepositEnsName]: {
    title: 'Validator by ENS of the deposit address',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 5,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalCredential]: {
    title: 'Validator by withdrawal credential',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalAddress]: {
    title: 'Validator by withdrawal address',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalEnsName]: {
    title: 'Validator by ENS of the withdrawal address',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 7,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByGraffiti]: {
    title: 'Validator by graffito',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByName]: {
    title: 'Validator by name',
    titleShort: 'Validator',
    logo: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false
  }
}

export interface SearchAheadSingleResult {
  chain_id: number,
  type: string,
  str_value?: string,
  num_value?: number,
  hash_value?: string
}

export interface SearchAheadResults {
  data?: SearchAheadSingleResult[],
  error?: string
}

// The parameter of the callback function that you give to SearchEngine.vue's props `pick-by-default` is an array of Matching elements. The function returns one Matching element.
export interface Matching {
  closeness: number, // if different results of this type exist in this network, only the best closeness is recorded here
  network: ChainIDs,
  type: ResultTypes
}

export function getListOfCategories () : Categories[] {
  const list : Categories[] = []

  for (const cat in Categories) {
    list.push(Categories[cat as keyof typeof Categories])
  }
  return list
}

export function getListOfResultTypes (sortByPriority : boolean) : ResultTypes[] {
  const list : ResultTypes[] = []

  for (const type in ResultTypes) {
    list.push(ResultTypes[type as keyof typeof ResultTypes])
  }
  if (sortByPriority) {
    list.sort((a, b) => { return TypeInfo[a].priority - TypeInfo[b].priority })
  }
  return list
}

const searchableTypesPerCategory : Record<string, ResultTypes[]> = {}
// Returns the list of types belonging to the given category.
// This function is fast on average: it computes the list only at the first call. Subsequent calls return the already computed list.
export function getListOfResultTypesInCategory (category: Categories) : ResultTypes[] {
  if (!searchableTypesPerCategory[category]) {
    const list : ResultTypes[] = []

    for (const t of getListOfResultTypes(true)) {
      if (TypeInfo[t].category === category) {
        list.push(t)
      }
    }
    searchableTypesPerCategory[category] = list
  }

  return searchableTypesPerCategory[category]
}

/*interface OrganizedSingleResult {
  columns: string[],
  suggestion: number,
  closeness: number
}
apiResponseElement
  str_value?: string,
  num_value?: number,
  hash_value?: string
  