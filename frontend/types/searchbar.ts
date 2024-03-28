import { ChainIDs } from '~/types/networks'

export enum SearchbarStyle {
  Gaudy = 'gaudy',
  Discreet = 'discreet',
  Embedded = 'embedded'
}
export enum SearchbarPurpose { General, Accounts, Validators }

export enum Category {
  Tokens,
  NFTs,
  Protocol,
  Addresses,
  Validators
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
   triggers the event `@go` to call your handler with the actual data of the result that you picked. If you return `undefined` instead
   of a matching, nothing happens (either no result suits you or you want to deactivate Enter).
   You will find futher below a function named pickHighestPriorityAmongMostRelevantMatchings. It is an example that you can use directly. */
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

// in SuggestionRow.vue, you will see that the drop-down where the list of result suggestions appear is organised into 3 rows that display a "name", a "description" and some "low level data", about each result
export type ResultSuggestionOutput = {
  name : string,
  description : string,
  lowLevelData : string
}

// The next 2 types will determine what data we must write into the differient fields of ResultSuggestionOutput after the API responded
export enum FillFrom {
  SASRstr_value,
  SASRnum_value,
  SASRhash_value,
  CategoryTitle,
  SubCategoryTitle,
  TypeTitle
}
export interface HowToFillresultSuggestionOutput {
  name : FillFrom | string,
  description : FillFrom | string,
  lowLevelData : FillFrom | string,
}

export interface ResultSuggestion {
  output: ResultSuggestionOutput,
  nameWasUnknown : boolean,
  queryParam: string, // data returned by the API that identifies this very result in the back-end (will be given to the callback function `@go`)
  closeness: number, // how close the suggested result is to the user input (important for graffiti, later for other things if the back-end evolves to find other approximate results)
  count : number, // How many identical results are found (often 1 but the API can inform us if there is more). This value is NaN when there is at least 1 result but the API did not clarify how many.
  rawResult: SearchAheadSingleResult // reference to the original data given by the API
}

export interface OrganizedResults {
  networks: {
    chainId: ChainIDs,
    types: {
      type: ResultType,
      suggestions: ResultSuggestion[]
    }[]
  }[]
}

interface SearchbarPurposeInfoField {
  searchable : Category[], // list of categories that the bar can search in
  unsearchable : ResultType[] // list of types that the bar will not search for
}
export const SearchbarPurposeInfo: Record<SearchbarPurpose, SearchbarPurposeInfoField> = {
  [SearchbarPurpose.General]: {
    searchable: [Category.Protocol, Category.Addresses, Category.Tokens, Category.NFTs, Category.Validators],
    unsearchable: []
  },
  [SearchbarPurpose.Accounts]: {
    searchable: [Category.Addresses],
    unsearchable: [ResultType.EnsOverview]
  },
  [SearchbarPurpose.Validators]: {
    searchable: [Category.Validators],
    unsearchable: []
  }
}

interface CategoryInfoFields {
  title : string,
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
  countable: boolean, // whether it is possible for the API to find several identical results of this type and count them
  queryParamField : FillFrom, // name of the field in SearchAheadSingleResult whose data identifies precisely the result in the back-end (this data will be passed to your `@go` call-back function when a result suggestion has been chosen)
  howToFillresultSuggestionOutput : HowToFillresultSuggestionOutput // will be used at execution time to know what data we must copy into each ResultSuggestionOutput
}

export const TypeInfo: Record<ResultType, TypeInfoFields> = {
  [ResultType.Tokens]: {
    title: 'ERC-20 token',
    category: Category.Tokens,
    subCategory: SubCategory.Tokens,
    priority: 3,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRstr_value, // this tells us that field `str_value` in SearchAheadSingleResult identifies precisely a result of type ResultType.Tokens when communicating about it with the back-end
    howToFillresultSuggestionOutput: { name: FillFrom.SASRstr_value, description: '', lowLevelData: FillFrom.SASRhash_value } // this tells us that field `name` in ResultSuggestionOutput will be filled with the content of `str_value` in SearchAheadSingleResult, and `lowLevelData` will be filled with `hash_value`
  },
  [ResultType.NFTs]: {
    title: 'ERC-721 & ERC-1155 token (NFT)',
    category: Category.NFTs,
    subCategory: SubCategory.NFTs,
    priority: 4,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SASRstr_value, description: '', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.Epochs]: {
    title: 'Epoch',
    category: Category.Protocol,
    subCategory: SubCategory.Epochs,
    priority: 12,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: '' }
  },
  [ResultType.Slots]: {
    title: 'Slot',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 11,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.Blocks]: {
    title: 'Block',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 10,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.BlockRoots]: {
    title: 'Block root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 18,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.StateRoots]: {
    title: 'State root',
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 19,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.Transactions]: {
    title: 'Transaction',
    category: Category.Protocol,
    subCategory: SubCategory.Transactions,
    priority: 17,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: '', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.TransactionBatches]: {
    title: 'Tx Batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 14,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: '' }
  },
  [ResultType.StateBatches]: {
    title: 'State batch',
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 13,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: FillFrom.SASRnum_value, lowLevelData: '' }
  },
  [ResultType.Contracts]: {
    title: 'Contract',
    category: Category.Addresses,
    subCategory: SubCategory.Contracts,
    priority: 2,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SASRstr_value, description: '', lowLevelData: FillFrom.SASRhash_value } // str_value is the name of the contract (for ex: "uniswap") but if the API gives '' we will replace it with a generic name (the title of this type: "Contract")
  },
  [ResultType.Accounts]: {
    title: 'Account',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 2,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: '', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.EnsAddresses]: {
    title: 'ENS address',
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 1,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SASRstr_value, description: '', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.EnsOverview]: {
    title: 'Overview of ENS domain',
    category: Category.Addresses,
    subCategory: SubCategory.EnsOverview,
    priority: 15,
    belongsToAllNetworks: true,
    countable: false,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: FillFrom.SASRstr_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.Graffiti]: {
    title: 'Graffiti',
    category: Category.Protocol,
    subCategory: SubCategory.Graffiti,
    priority: 16,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.TypeTitle, description: 'Blocks with', lowLevelData: FillFrom.SASRstr_value }
  },
  [ResultType.ValidatorsByIndex]: {
    title: 'Validator by index',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRnum_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.ValidatorsByPubkey]: {
    title: 'Validator by public key',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countable: false,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: FillFrom.SASRnum_value, lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.ValidatorsByDepositAddress]: {
    title: 'Validator by deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 6,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Deposited by', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.ValidatorsByDepositEnsName]: {
    title: 'Validator by ENS of the deposit address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 5,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Deposited by', lowLevelData: FillFrom.SASRstr_value }
  },
  [ResultType.ValidatorsByWithdrawalCredential]: {
    title: 'Validator by withdrawal credential',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Credential', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.ValidatorsByWithdrawalAddress]: {
    title: 'Validator by withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRhash_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Withdrawn to', lowLevelData: FillFrom.SASRhash_value }
  },
  [ResultType.ValidatorsByWithdrawalEnsName]: {
    title: 'Validator by ENS of the withdrawal address',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 7,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Withdrawn to', lowLevelData: FillFrom.SASRstr_value }
  },
  [ResultType.ValidatorsByGraffiti]: {
    title: 'Validator by graffito',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Block graffiti', lowLevelData: FillFrom.SASRstr_value }
  },
  [ResultType.ValidatorsByName]: {
    title: 'Validator by name',
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countable: true,
    queryParamField: FillFrom.SASRstr_value,
    howToFillresultSuggestionOutput: { name: FillFrom.SubCategoryTitle, description: 'Named', lowLevelData: FillFrom.SASRstr_value }
  }
}

export function isOutputAnAPIresponse (type : ResultType, resultSuggestionOutputField : keyof HowToFillresultSuggestionOutput) : boolean {
  switch (TypeInfo[type].howToFillresultSuggestionOutput[resultSuggestionOutputField]) {
    case FillFrom.SASRstr_value :
    case FillFrom.SASRnum_value :
    case FillFrom.SASRhash_value :
      return true
    default:
      return false
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

// This is an example of function that <BcSearchbarMain> needs in its props `pick-by-default`. You can design a function fulfilling your needs
// or simply give this one (after importing pickHighestPriorityAmongMostRelevantMatchings at the top of your script setup) if it does what you
// need.
// The purpose of the function given to props `pick-by-default` is to pick a result when the user presses Enter instead of clicking a result in
// the drop-down. Note: if your function returns `undefined` it means that either no result suits you or you want to deactivate Enter.
export function pickHighestPriorityAmongMostRelevantMatchings (possibilities : Matching[]) : Matching|undefined {
  // What this funtion works with:
  //   `possibilities` contains an abstract representation of the possible results sorted by network and type priority (the order appearing in
  //   the drop-down). We must select one of them or we can return `undefined` if no result suits us or if we want to deactivate Enter.
  // What we implemented in this example function:
  //   We look for the possibility that matches the best with the user input (this is known through the field `Matching.closeness`).
  //   If several possibilities with this best closeness value exist, we catch the first one (so the one having the highest priority). This
  //   happens for example when the user input corresponds to both a validator index and a block number.
  let bestMatchWithHigherPriority = possibilities[0]
  for (const possibility of possibilities) {
    if (possibility.closeness < bestMatchWithHigherPriority.closeness) {
      bestMatchWithHigherPriority = possibility
    }
  }
  return bestMatchWithHigherPriority
}
