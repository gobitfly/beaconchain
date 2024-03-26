import { ChainIDs } from '~/types/networks'

export type SearchBarStyle = 'discreet' | 'gaudy' | 'embedded'
export enum SearchBarPurpose { General, Accounts, Validators }

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
  EnsOverview,
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

// The parameter of the callback function that you give to <BcSearchbarMain>'s props `pick-by-default` is an array of Matching elements
// and the function returns one Matching element.
export interface Matching {
  closeness: number, // how close this result is to what the user inputted (lower value = better similarity)
  network: ChainIDs, // the network that this result belongs to
  type: ResultType // the type of the result
}
/* When the user presses Enter, the callback function receives a simplified representation of the suggested results and returns one
   element from this list (or undefined). This list is passed in parameter `possibilities` as a simplified view of the actual list of
   results. It is sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. After you return a matching, the bar
   triggers the event `@go` to call your handler with the actual data of the result that you picked. If you return undefined instead
   of a matching, nothing happens (either no result suits you or you want to deactivate Enter). */
export interface PickingCallBackFunction { (possibilities : Matching[]) : Matching|undefined }

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

export enum PredefinedFilling {
  CategoryTitle,
  SubCategoryTitle,
  TypeTitle
}

export type DropdownOutput = {
  name : string,
  description : string,
  lowLevelData : string
}

type DropdownOutputBeforeFilling = {
  name : undefined | string | PredefinedFilling,
  description : undefined | string | PredefinedFilling,
  lowLevelData : undefined | string | PredefinedFilling
}

export interface ResultSuggestion {
  output: DropdownOutput,
  nameWasUnknown : boolean,
  queryParam: string, // data returned by the API that identifies this very result in the back-end (will be given to the callback function `@go`)
  closeness: number // how close the suggested result is to the user input (important for graffiti, later for other things if the back-end evolves to find other approximate results)
  count : number // How many identical results are found (often 1 but the API can inform us if there is more). This value is NaN when there is at least 1 result but the API did not clarify how many.
}
export interface OrganizedResults {
  networks: {
    chainId: ChainIDs,
    types: {
      type: ResultType,
      suggestion: ResultSuggestion[]
    }[]
  }[]
}

interface CategoryInfoFields {
  title : string
  filterLabel : string
}

export const CategoryInfo: Record<Category, CategoryInfoFields> = {
  [Category.Tokens]: { title: 'ERC-20 Tokens', filterLabel: 'Tokens' },
  [Category.NFTs]: { title: 'NFTs', filterLabel: 'NFTs' },
  [Category.Protocol]: { title: 'Protocol', filterLabel: 'Protocol' },
  [Category.Addresses]: { title: 'Addresses', filterLabel: 'Addresses' },
  [Category.Validators]: { title: 'Validators', filterLabel: 'Validators' }
}

interface SubCategoryInfoFields {
  title : string
}

export const SubCategoryInfo: Record<SubCategory, SubCategoryInfoFields> = {
  [SubCategory.Tokens]: { title: 'Token' },
  [SubCategory.NFTs]: { title: 'NFT' },
  [SubCategory.Epochs]: { title: 'Epoch' },
  [SubCategory.SlotsAndBlocks]: { title: 'Slot/Block' },
  [SubCategory.Transactions]: { title: 'Transaction' },
  [SubCategory.Batches]: { title: 'Batch' },
  [SubCategory.Contracts]: { title: 'Contract' },
  [SubCategory.Accounts]: { title: 'Account' },
  [SubCategory.EnsOverview]: { title: 'ENS Overview' },
  [SubCategory.Graffiti]: { title: 'Graffiti' },
  [SubCategory.Validators]: { title: 'Validator' }
}

interface TypeInfoFields {
  title: string,
  category: Category,
  subCategory: SubCategory,
  priority: number,
  belongsToAllNetworks: boolean,
  countable: boolean, // whether it is possible for the API to find several identical results and count them
  fieldsInSearchAheadResult : (keyof SearchAheadSingleResult)[], // fields to read from the SearchAheadSingleResult object returned by the API. The order of these field-names sets the order that we use to fill the undefined data of dropdownOutputBeforeFilling (see just below here)
  queryParamField : keyof SearchAheadSingleResult, // name of the field in SearchAheadSingleResult whose data identifies precisely the result in the back-end
  dropdownOutputBeforeFilling : DropdownOutputBeforeFilling // Information to show when a result of this type is suggested in the drop-down. As instructed by isOutputAnAPIresponse() (see further down below), undefined elements will be filled during execution from the fields named in fieldsInSearchAheadResult (see just above here).
}

export const TypeInfo: Record<ResultType, TypeInfoFields> = {
  [ResultType.Tokens]: {
    title: 'ERC-20 token',
    category: Category.Tokens,
    subCategory: SubCategory.Tokens,
    priority: 3,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['str_value', 'hash_value'], // This means that we read the token name and the token address from the API response and fill the undefined fields of ResultSuggestion.output` (see just below) with this information in that order.
    queryParamField: 'str_value', // This is the name of the field in SearchAheadSingleResult which identifies precisely a result when communicating with the back-end.
    dropdownOutputBeforeFilling: { name: undefined, description: '', lowLevelData: undefined } // These `undefined`s will be replaced during execution with the data named in fieldsInSearchAheadResult (see just above), respectively str_value and hash_value in that order.
  },
  [ResultType.NFTs]: {
    title: 'ERC-721 & ERC-1155 token (NFT)',
    category: Category.NFTs,
    subCategory: SubCategory.NFTs,
    priority: 4,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['str_value', 'hash_value'], // token name, token address
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: undefined, description: '', lowLevelData: undefined }
  },
  [ResultType.Epochs]: {
    title: 'Epoch',
    category: Category.Protocol,
    subCategory: SubCategory.Epochs,
    priority: 12,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value'],
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: '' }
  },
  [ResultType.Slots]: {
    title: 'Slot',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 11,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'], // num_value is the slot number, hash_value is the state root if it is what the user typed otherwise it contains by default the block root
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.Blocks]: {
    title: 'Block',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 10,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'], // same as above
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.BlockRoots]: {
    title: 'Block root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 18,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.StateRoots]: {
    title: 'State root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 19,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'],
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.Transactions]: {
    title: 'Transaction',
    category: Category.Protocol,
    subCategory: SubCategory.Transactions,
    priority: 17,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['hash_value'],
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: '', lowLevelData: undefined }
  },
  [ResultType.TransactionBatches]: {
    title: 'TX Batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 14,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value'],
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: '' }
  },
  [ResultType.StateBatches]: {
    title: 'State batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 13,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value'],
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: undefined, lowLevelData: '' }
  },
  [ResultType.Contracts]: {
    title: 'Contract',
    category: Category.Addresses,
    subCategory: SubCategory.Contracts,
    priority: 2,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['str_value', 'hash_value'], // str_value is the name of the contract  (for ex: "uniswap") or "" by default if unknown
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: undefined, description: '', lowLevelData: undefined } // if the API gives '' for the first element (0), we will replace it with a generic name (the title of this type: "Contract")
  },
  [ResultType.Accounts]: {
    title: 'Account',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 2,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['hash_value'],
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: '', lowLevelData: undefined }
  },
  [ResultType.EnsAddresses]: {
    title: 'ENS address',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 1,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['str_value', 'hash_value'], // ENS name, corresponding address
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: undefined, description: '', lowLevelData: undefined }
  },
  [ResultType.EnsOverview]: {
    title: 'Overview of ENS domain',
    category: Category.Addresses,
    subCategory: SubCategory.EnsOverview,
    priority: 15,
    belongsToAllNetworks: true,
    countable: false,
    fieldsInSearchAheadResult: ['str_value', 'hash_value'], // same as above
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.Graffiti]: {
    title: 'Graffito',
    category: Category.Protocol,
    subCategory: SubCategory.Graffiti,
    priority: 16,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['str_value'],
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.TypeTitle, description: 'Blocks with', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByIndex]: {
    title: 'Validator by index',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamField: 'num_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.ValidatorsByPubkey]: {
    title: 'Validator by public key',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countable: false,
    fieldsInSearchAheadResult: ['num_value', 'hash_value'], // validator index, pubkey
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: undefined, lowLevelData: undefined }
  },
  [ResultType.ValidatorsByDepositAddress]: {
    title: 'Validator by deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 6,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['hash_value'], // deposit address
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Deposited by', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByDepositEnsName]: {
    title: 'Validator by ENS of the deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 5,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['str_value'], // ENS name
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Deposited by', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByWithdrawalCredential]: {
    title: 'Validator by withdrawal credential',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['hash_value'], // withdrawal credential
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Credential', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByWithdrawalAddress]: {
    title: 'Validator by withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['hash_value'], // withdrawal address
    queryParamField: 'hash_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Withdrawn to', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByWithdrawalEnsName]: {
    title: 'Validator by ENS of the withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 7,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['str_value'], // ENS name
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Withdrawn to', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByGraffiti]: {
    title: 'Validator by graffito',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['str_value'], // graffito
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Block graffiti', lowLevelData: undefined }
  },
  [ResultType.ValidatorsByName]: {
    title: 'Validator by name',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countable: true,
    fieldsInSearchAheadResult: ['str_value'], // name that the owner recorded on beaconcha.in
    queryParamField: 'str_value',
    dropdownOutputBeforeFilling: { name: PredefinedFilling.SubCategoryTitle, description: 'Named', lowLevelData: undefined }
  }
}

export function isOutputAnAPIresponse (type : ResultType, dropdownOutputField : keyof DropdownOutput) : boolean {
  return TypeInfo[type].dropdownOutputBeforeFilling[dropdownOutputField] === undefined
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
