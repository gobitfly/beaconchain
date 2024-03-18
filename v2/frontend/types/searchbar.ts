import { ChainIDs } from '~/types/networks'

export type SearchBarStyle = 'discreet' | 'gaudy' | 'embedded'

export enum Category {
  Tokens = 'tokens',
  NFTs = 'nfts',
  Protocol = 'protocol',
  Addresses = 'addresses',
  Validators = 'validators'
}

export enum SubCategory {
  Tokens,
  NFTs,
  Epochs,
  SlotsAndBlocks,
  Transactions,
  Batches,
  Contracts,
  Accounts,
  EnsSystem,
  Graffiti,
  Validators
}

export enum ResultType {
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
  Contracts = 'contracts',
  Accounts = 'accounts',
  EnsAddresses = 'ens_addresses',
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

export const CategoryInfo: Record<Category, CategoryInfoFields> = {
  [Category.Tokens]: { filterLabel: 'Tokens' },
  [Category.NFTs]: { filterLabel: 'NFTs' },
  [Category.Protocol]: { filterLabel: 'Protocol' },
  [Category.Addresses]: { filterLabel: 'Addresses' },
  [Category.Validators]: { filterLabel: 'Validators' }
}

// The parameter of the callback function that you give to <BcSearchbarMainComponent>'s props `pick-by-default` is an array of Matching elements. The function returns one Matching element.
export interface Matching {
  closeness: number, // if different results of this type exist on the network, only the best closeness is recorded here
  network: ChainIDs,
  type: ResultType
}

export interface SearchAheadSingleResult {
  chain_id: number,
  type: string,
  str_value?: string,
  num_value?: number,
  hash_value?: string
}

export interface SearchAheadResult {
  data?: SearchAheadSingleResult[],
  error?: string
}

interface TypeInfoFields {
  title: string,
  category: Category,
  subCategory: SubCategory,
  priority: number,
  belongsToAllNetworks: boolean,
  dataInSearchAheadResult : (keyof SearchAheadSingleResult)[], // the order of these field-names sets the order of the information displayed in the dropdown
  queryParamIndex : number, // points to the field-name in array `dataInSearchAheadResult` whose data will be understood by the back-end as a reference to a result
  dropdownColumns : (string|undefined)[] // Information to show when a result of this type is suggested in the drop-down. The undefined elements will be filled during execution with the fields given just above here (dataInSearchAheadResult). The first column often names the type. If so, it can be set to '' statically, which will be replaced during execution with the content of field `title` above.
}

export const TypeInfo: Record<ResultType, TypeInfoFields> = {
  [ResultType.Tokens]: {
    title: 'ERC-20 token',
    category: Category.Tokens,
    subCategory: SubCategory.Tokens,
    priority: 3,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // This means that we read the token name and the token address from the API response and will fill array `dropdownColumns` (see below) with this information in that order.
    queryParamIndex: 0,
    dropdownColumns: [undefined, '', undefined] // These `undefined`s will be replaced during execution with what is given above here, respectively str_value and hash_value in that order. So the first information displayed in the drop-down will be a string, the second info will be a hash. According to '', the last column of information will be left empty.
  },
  [ResultType.NFTs]: {
    title: 'ERC-721 & ERC-1155 token (NFT)',
    category: Category.NFTs,
    subCategory: SubCategory.NFTs,
    priority: 4,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // token name, token address
    queryParamIndex: 0,
    dropdownColumns: [undefined, '', undefined]
  },
  [ResultType.Epochs]: {
    title: 'Epoch',
    category: Category.Protocol,
    subCategory: SubCategory.Epochs,
    priority: 12,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultType.Slots]: {
    title: 'Slot',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 11,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // num_value is the slot number, hash_value is the state root if it is what the user typed otherwise it contains by default the block root
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultType.Blocks]: {
    title: 'Block',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 10,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // same as above
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultType.BlockRoots]: {
    title: 'Block root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 18,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultType.StateRoots]: {
    title: 'State root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 19,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, undefined]
  },
  [ResultType.Transactions]: {
    title: 'Transaction',
    category: Category.Protocol,
    subCategory: SubCategory.Transactions,
    priority: 17,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', '', undefined]
  },
  [ResultType.TransactionBatches]: {
    title: 'Transaction batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 14,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['TX Batch', undefined, '']
  },
  [ResultType.StateBatches]: {
    title: 'State batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 13,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', undefined, '']
  },
  [ResultType.Contracts]: {
    title: 'Contract',
    category: Category.Addresses,
    subCategory: SubCategory.Contracts,
    priority: 2,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // str_value is the name of the contract  (for ex: "uniswap") or "" by default if unknown
    queryParamIndex: 1,
    dropdownColumns: [undefined, '', undefined]
  },
  [ResultType.Accounts]: {
    title: 'Account',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 2,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['hash_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', '', undefined]
  },
  [ResultType.EnsAddresses]: {
    title: 'ENS address',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 1,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // ENS name, corresponding address
    queryParamIndex: 0,
    dropdownColumns: [undefined, '', undefined]
  },
  [ResultType.EnsOverview]: {
    title: 'Overview of ENS domain',
    category: Category.Addresses,
    subCategory: SubCategory.EnsSystem,
    priority: 15,
    belongsToAllNetworks: true,
    dataInSearchAheadResult: ['str_value', 'hash_value'], // same as above
    queryParamIndex: 0,
    dropdownColumns: ['ENS Overview', undefined, undefined]
  },
  [ResultType.Graffiti]: {
    title: 'Graffito',
    category: Category.Protocol,
    subCategory: SubCategory.Graffiti,
    priority: 16,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'],
    queryParamIndex: 0,
    dropdownColumns: ['', 'Blocks with', undefined]
  },
  [ResultType.ValidatorsByIndex]: {
    title: 'Validator by index',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamIndex: 0,
    dropdownColumns: ['Validator', undefined, undefined]
  },
  [ResultType.ValidatorsByPubkey]: {
    title: 'Validator by public key',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamIndex: 1,
    dropdownColumns: ['Validator', undefined, undefined]
  },
  [ResultType.ValidatorsByDepositAddress]: {
    title: 'Validator by deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 6,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // deposit address
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Deposited by', undefined]
  },
  [ResultType.ValidatorsByDepositEnsName]: {
    title: 'Validator by ENS of the deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 5,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // ENS name
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Deposited by', undefined]
  },
  [ResultType.ValidatorsByWithdrawalCredential]: {
    title: 'Validator by withdrawal credential',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // withdrawal credential
    queryParamIndex: 0,
    dropdownColumns: ['Validator', '', undefined]
  },
  [ResultType.ValidatorsByWithdrawalAddress]: {
    title: 'Validator by withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['hash_value'], // withdrawal address
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Withdrawn to', undefined]
  },
  [ResultType.ValidatorsByWithdrawalEnsName]: {
    title: 'Validator by ENS of the withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 7,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // ENS name
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Withdrawn to', undefined]
  },
  [ResultType.ValidatorsByGraffiti]: {
    title: 'Validator by graffito',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // graffito
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Tagging with', undefined]
  },
  [ResultType.ValidatorsByName]: {
    title: 'Validator by name',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    dataInSearchAheadResult: ['str_value'], // name that the owner recorded on beaconcha.in
    queryParamIndex: 0,
    dropdownColumns: ['Validator', 'Named', undefined]
  }
}

export function getListOfCategories () : Category[] {
  const list : Category[] = []

  for (const cat in Category) {
    list.push(Category[cat as keyof typeof Category])
  }
  return list
}

const listOfResultTypesAsDeclared : ResultType[] = []
const listOfResultTypesPrioritized : ResultType[] = []
// Returns all litterals in `ResultType` used to communicate with the API.
// This function is fast on average: it computes the list only at the first call. Subsequent calls return the already computed list.
export function getListOfResultTypes (sortByPriority : boolean) : ResultType[] {
  if (listOfResultTypesAsDeclared.length === 0) {
    for (const type in ResultType) {
      const ty = type as keyof typeof ResultType
      listOfResultTypesAsDeclared.push(ResultType[ty])
      listOfResultTypesPrioritized.push(ResultType[ty])
    }
    listOfResultTypesPrioritized.sort((a, b) => { return TypeInfo[a].priority - TypeInfo[b].priority })
  }

  return sortByPriority ? listOfResultTypesPrioritized : listOfResultTypesAsDeclared
}

const searchableTypesPerCategory : Record<string, ResultType[]> = {}
// Returns the list of types belonging to the given category.
// This function is fast on average: it computes the lists only at the first call. Subsequent calls return the already computed lists.
export function getListOfResultTypesInCategory (category: Category, sortByPriority : boolean) : ResultType[] {
  if (!(category in searchableTypesPerCategory)) {
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
