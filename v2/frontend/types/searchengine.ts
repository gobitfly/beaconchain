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

// The parameter of the callback function that you give to SearchEngine.vue's props `pick-by-default` is an array of Matching elements. The function returns one Matching element.
export interface Matching {
  closeness: number, // if different results of this type exist on the network, only the best closeness is recorded here
  network: ChainIDs,
  type: ResultTypes
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

interface TypeInfoFields {
  title: string,
  logo: string,
  category: Categories,
  priority: number,
  belongsToAllNetworks: boolean,
  dataInSearchAheadResult : (keyof SearchAheadSingleResult)[], // the order of these field-names sets the order of the information displayed in the dropdown
  queryParamIndex : number, // points to the field-name in the array above whose data will be understood by the back-end as a reference to a result
  dropdownColumns : (string|undefined)[] // Static information to show when a result of this type is suggested in the drop-down. The undefined elements will be filled during execution with the fields given just above here (dataInSearchAheadResult). The first column often names the type. If so, it can be set to '' statically, which will be replaced during execution with the content of field `title` (see above).
}

export const TypeInfo: Record<ResultTypes, TypeInfoFields> = {
  [ResultTypes.Tokens]: {
    title: 'ERC-20 token',
    logo: '',
    category: Categories.Tokens,
    priority: 3,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // token name, token address
    queryParamIndex: 0,
    dropdownColumns: [undefined, undefined, '']
  },
  [ResultTypes.NFTs]: {
    title: 'NFT (ERC-721 & ERC-1155 token)',
    logo: '',
    category: Categories.NFTs,
    priority: 4,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // token name, token address
    queryParamIndex: 0,
    dropdownColumns: [undefined, undefined, '']
  },
  [ResultTypes.Epochs]: {
    title: 'Epoch',
    logo: '',
    category: Categories.Protocol,
    priority: 12,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultTypes.Slots]: {
    title: 'Slot',
    logo: '',
    category: Categories.Protocol,
    priority: 11,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // num_value is the slot number, hash_value is the state root if it is what the user typed otherwise it contains by default the block root
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultTypes.Blocks]: {
    title: 'Block',
    logo: '',
    category: Categories.Protocol,
    priority: 10,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // same as above
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultTypes.BlockRoots]: {
    title: 'Block root',
    logo: '',
    category: Categories.Protocol,
    priority: 18,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultTypes.StateRoots]: {
    title: 'State root',
    logo: '',
    category: Categories.Protocol,
    priority: 19,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultTypes.Transactions]: {
    title: 'Transaction',
    logo: '',
    category: Categories.Protocol,
    priority: 17,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultTypes.TransactionBatches]: {
    title: 'Transaction batch',
    logo: '',
    category: Categories.Protocol,
    priority: 14,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['TX Batch', undefined, '']
  },
  [ResultTypes.StateBatches]: {
    title: 'State batch',
    logo: '',
    category: Categories.Protocol,
    priority: 13,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultTypes.Accounts]: {
    title: 'Account',
    logo: '',
    category: Categories.Addresses,
    priority: 2,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultTypes.Contracts]: {
    title: 'Contract',
    logo: '',
    category: Categories.Addresses,
    priority: 2,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // str_value is the name of the contract  (for ex: "uniswap") or "" by default if unknown
    queryParamIndex: 1,
    dropdownColumns: [undefined, undefined, '']
  },
  [ResultTypes.Ens]: {
    title: 'ENS address',
    logo: '',
    category: Categories.Addresses,
    priority: 1,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // ENS name, corresponding address
    queryParamIndex: 0,
    dropdownColumns: [undefined, undefined, '']
  },
  [ResultTypes.EnsOverview]: {
    title: 'Overview of ENS domain',
    logo: '',
    category: Categories.Addresses,
    priority: 15,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // same as above
    queryParamIndex: 0,
    dropdownColumns: ['ENS Overview', undefined, undefined]
  },
  [ResultTypes.Graffiti]: {
    title: 'Graffito',
    logo: '',
    category: Categories.Protocol,
    priority: 16,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', 'Blocks with', undefined]
  },
  [ResultTypes.ValidatorsByIndex]: {
    title: 'Validator by index',
    logo: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamIndex: 0,
    dropdownColumns: ['Validator', undefined, undefined]
  },
  [ResultTypes.ValidatorsByPubkey]: {
    title: 'Validator by public key',
    logo: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamIndex: 1,
    dropdownColumns: ['Validator', undefined, undefined]
  },
  [ResultTypes.ValidatorsByDepositAddress]: {
    title: 'Validator by deposit address',
    logo: '',
    category: Categories.Validators,
    priority: 6,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // deposit address
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Deposited by', undefined]
  },
  [ResultTypes.ValidatorsByDepositEnsName]: {
    title: 'Validator by ENS of the deposit address',
    logo: '',
    category: Categories.Validators,
    priority: 5,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // ENS name
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Deposited by', undefined]
  },
  [ResultTypes.ValidatorsByWithdrawalCredential]: {
    title: 'Validator by withdrawal credential',
    logo: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // withdrawal credential
    queryParamIndex: 0,
    dropdownColumns: ['Validator', undefined, '']
  },
  [ResultTypes.ValidatorsByWithdrawalAddress]: {
    title: 'Validator by withdrawal address',
    logo: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // withdrawal address
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Withdrawn to', undefined]
  },
  [ResultTypes.ValidatorsByWithdrawalEnsName]: {
    title: 'Validator by ENS of the withdrawal address',
    logo: '',
    category: Categories.Validators,
    priority: 7,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // ENS name
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Withdrawn to', undefined]
  },
  [ResultTypes.ValidatorsByGraffiti]: {
    title: 'Validator by graffito',
    logo: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // graffito
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Tagging with', undefined]
  },
  [ResultTypes.ValidatorsByName]: {
    title: 'Validator by name',
    logo: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // name that the owner recorded on beaconcha.in
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Named', undefined]
  }
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
// This function is fast on average: it computes the lists only at the first call. Subsequent calls return the already computed lists.
export function getListOfResultTypesInCategory (category: Categories, sortByPriority : boolean) : ResultTypes[] {
  if (Object.keys(searchableTypesPerCategory).length === 0) {
    for (const t of getListOfResultTypes(sortByPriority)) {
      const c = TypeInfo[t].category
      if (!searchableTypesPerCategory[c]) {
        searchableTypesPerCategory[c] = []
      }
      searchableTypesPerCategory[c].push(t)
    }
  }

  return searchableTypesPerCategory[category]
}
