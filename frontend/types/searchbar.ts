import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'
import { ChainIDs } from '~/types/networks'
import { type ApiErrorResponse, type SearchResult, type InternalPostSearchResponse } from '~/types/api/common'

export enum SearchbarShape { // do not change the litterals, they are used as class names
  Small = 'small',
  Medium = 'medium',
  Big = 'big'
}
export enum SearchbarColors { // do not change the litterals, they are used as class names
  Default = 'default',
  DarkBlue = 'darkblue',
  LightBlue = 'lightblue'
}
export enum SearchbarPurpose {
  GlobalSearch,
  AccountAddition,
  ValidatorAddition
}

// List of the possible ways to organise the information in each row of the result-suggestion list.
export enum SuggestionrowCells {
  NameDescriptionLowlevelCategory, // this mode displays everything
  SubcategoryIdentificationDescription // shows less information, and this information is a bit different from what the complete mode shows
  // If you add here a display-mode, update the `SearchbarPurposeInfo` record to tell SuggestionRow.vue which display-mode
  // it should use, and then implement this mode in the `<template>` of SuggestionRow.vue.
}

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
  ValidatorsByIndex = 'validator_by_index',
  ValidatorsByPubkey = 'validator_by_public_key',
  ValidatorsByDepositAddress = 'validators_by_deposit_address',
  ValidatorsByDepositEnsName = 'validators_by_deposit_ens_name',
  ValidatorsByWithdrawalCredential = 'validators_by_withdrawal_credential',
  ValidatorsByWithdrawalAddress = 'validators_by_withdrawal_address',
  ValidatorsByWithdrawalEnsName = 'validators_by_withdrawal_ens_name',
  ValidatorsByGraffiti = 'validators_by_graffiti'
  // ValidatorsByName = 'validators_by_name'  // for users having given a name to their validator in our DB
}

// The parameter of the callback function that you give to <BcSearchbarMain>'s props `pick-by-default` is an array of `Matching` elements
// and the function returns one `Matching` element (or undefined).
export type Matching = {
  closeness: number, // how close this result is to what the user inputted (lower value = better similarity)
  network: ChainIDs, // the network that this result belongs to
  type: ResultType // the type of the result
}
/* When the user presses Enter, the callback function receives a simplified representation of the suggested results and returns one
   element from this list (or undefined). This list is passed in parameter `possibilities` as a simplified view of the actual list of
   results. It is sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. After you return a matching, the bar
   triggers the event `@go` to call your handler with the actual data of the matching that you picked. If you return `undefined` instead
   of a matching, nothing happens (either no result suits you or you want to deactivate Enter).
   You will find futher below a function named `pickHighestPriorityAmongBestMatchings`. It is an example that you can use directly. */
export interface PickingCallBackFunction { (possibilities : Matching[]) : Matching|undefined }

export interface SearchRequest {
  input: string,
  networks: ChainIDs[],
  types: ResultType[],
  count?: boolean
}
export type SingleAPIresult = SearchResult
export interface SearchAheadAPIresponse extends ApiErrorResponse, InternalPostSearchResponse {}

// in SuggestionRow.vue, you will see that the drop-down where the list of result suggestions appear is organised into 3 rows that display a "name", a "description" and some "low level data", about each result
export type ResultSuggestionOutput = {
  name : string,
  description : string,
  lowLevelData : string
}

// This type determines different sources that we can retrieve data from, mainly to fill the fields of ResultSuggestionOutput after the API responded
export enum Indirect {
  None = 0,
  APIstr_value,
  APInum_value,
  APIhash_value,
  CategoryTitle,
  SubCategoryTitle,
  TypeTitle
}
// The following 3 definitions will be used as parameters of function `t()` of I18n.
export type TranslatableLitteral = [string, number] // you will have to destructure the parameters with an ellipsis, like so: t(...myTranslatableLitteral)
const SINGULAR = 1
const PLURAL = 2 // Any number greater than 1 is good, this is just for I18n to show the plural form of the litteral constants that we define through the rest of the file.
// Hint: if you need to get the path of a TranslatableLitteral to give it to I18n (typically to change a singular into plural or vice-versa), use our function getI18nPathOfTranslatableLitteral() defined further below
// this type determines all the possible ways to fill the fields of ResultSuggestionOutput after the API responded
export type FillFrom = Indirect | TranslatableLitteral | ''

export interface HowToFillresultSuggestionOutput {
  name : FillFrom,
  description : FillFrom,
  lowLevelData : FillFrom,
}

export interface ResultSuggestion {
  output: ResultSuggestionOutput,
  queryParam: string, // Data returned by the API that identifies this very result in the back-end. This is the most important data for callback function '@go' given in the props of the Searchbar component.
  closeness: number, // How close the suggested result is to the user input (important for graffitis and token names, later for other things if the back-end evolves to find other approximate results).
  count : number, // How many identical results are found (often 1 but the API can inform us if there is more). This value is NaN when there is at least 1 result but the API did not clarify how many.
  chainId : ChainIDs, // Network that the result belongs to. If the result exists on all networks, it is `ChainIDs.Any` (so 0).
  type : ResultType, // Tells what thing(s) this result corresponds to.
  rawResult: SingleAPIresult // Original data given by the API.
}

export interface ResultSuggestionInternal extends ResultSuggestion {
  stringifyiedRawResult : string, // Original data given by the API.
  nameWasUnknown : boolean, // Tells whether the API had the possibility to fill field `name` in `output` but could not.
  lacksPremiumSubscription : boolean // `true` if the result is not accessible to the user due to account restrictions
}

export interface OrganizedResults {
  networks: {
    chainId: ChainIDs,
    types: {
      type: ResultType,
      suggestions: ResultSuggestionInternal[]
    }[]
  }[]
}

interface SearchbarPurposeInfoField {
  searchable: Category[], // List of categories that the bar can search in. The cateogry filter-buttons will appear on the screen in the same order as in this list.
  unsearchable: ResultType[], // List of types that the bar will not search for.
  askAPItoCountResults: boolean, // If `true`, the search-bar will ask the API explicitely to count results when what it searches for can be counted (this is told by field `countable` in the TypeInfo record further below). Note that even if not asked, the API can still return counts and batches and we will read this information anyway if so.
  button: 'search' | 'add', // Utility of the button.
  placeHolder: string, // I18n path of the hint to display in the input field when it is empty.
  cellsInSuggestionRows: SuggestionrowCells, // Determines what is shown in each row of the result-suggestion list.
  differentialRequests: boolean // If activated, the bar decreases the workload for the API **in certain scenarii** by asking only for results that it does not know yet (which can happen when the user started a search with filters and activates a new filter, then the bar asks only for results corresponding to the newly selected filter). The downside is that the bar cannot help the user by mentionning the number of filtered-out results at the bottom of the suggestion list.
}
// this Record describes the look and behavior of the search-bar according to the value that you pass in its props `:bar-purpose`
export const SearchbarPurposeInfo: Record<SearchbarPurpose, SearchbarPurposeInfoField> = {
  [SearchbarPurpose.GlobalSearch]: {
    searchable: [Category.Protocol, Category.Addresses, Category.Tokens, Category.NFTs, Category.Validators], // to display the filter buttons in a different order, write the categories in a different order here
    unsearchable: [],
    askAPItoCountResults: false,
    button: 'search',
    placeHolder: 'search_bar.general_placeholder',
    cellsInSuggestionRows: SuggestionrowCells.NameDescriptionLowlevelCategory,
    differentialRequests: true
  },
  [SearchbarPurpose.AccountAddition]: {
    searchable: [Category.Addresses],
    unsearchable: [ResultType.EnsOverview],
    askAPItoCountResults: false,
    button: 'add',
    placeHolder: 'search_bar.account_placeholder',
    cellsInSuggestionRows: SuggestionrowCells.SubcategoryIdentificationDescription,
    differentialRequests: true
  },
  [SearchbarPurpose.ValidatorAddition]: {
    searchable: [Category.Validators],
    unsearchable: [],
    askAPItoCountResults: false,
    button: 'add',
    placeHolder: 'search_bar.validator_placeholder',
    cellsInSuggestionRows: SuggestionrowCells.SubcategoryIdentificationDescription,
    differentialRequests: true
  }
}

interface CategoryInfoFields {
  title : TranslatableLitteral,
  filterLabel : TranslatableLitteral
}
export const CategoryInfo: Record<Category, CategoryInfoFields> = {
  [Category.Tokens]: { title: ['common.erc20token', PLURAL], filterLabel: ['common.token', PLURAL] },
  [Category.NFTs]: { title: ['common.nft', PLURAL], filterLabel: ['common.nft', PLURAL] },
  [Category.Protocol]: { title: ['common.protocol', SINGULAR], filterLabel: ['common.protocol', SINGULAR] },
  [Category.Addresses]: { title: ['common.address', PLURAL], filterLabel: ['common.address', PLURAL] },
  [Category.Validators]: { title: ['common.validator', PLURAL], filterLabel: ['common.validator', PLURAL] }
}

interface SubCategoryInfoFields {
  title : TranslatableLitteral
}
export const SubCategoryInfo: Record<SubCategory, SubCategoryInfoFields> = {
  [SubCategory.Tokens]: { title: ['common.token', SINGULAR] },
  [SubCategory.NFTs]: { title: ['common.nft', SINGULAR] },
  [SubCategory.Epochs]: { title: ['common.epoch', SINGULAR] },
  [SubCategory.SlotsAndBlocks]: { title: ['common.slot_block', SINGULAR] },
  [SubCategory.Transactions]: { title: ['common.transaction', SINGULAR] },
  [SubCategory.Batches]: { title: ['common.batch', SINGULAR] },
  [SubCategory.Contracts]: { title: ['common.contract', SINGULAR] },
  [SubCategory.Accounts]: { title: ['common.account', SINGULAR] },
  [SubCategory.EnsOverview]: { title: ['search_bar.ens_overview', SINGULAR] },
  [SubCategory.Graffiti]: { title: ['common.graffiti', SINGULAR] },
  [SubCategory.Validators]: { title: ['common.validator', SINGULAR] }
}

interface TypeInfoFields {
  title: TranslatableLitteral,
  category: Category,
  subCategory: SubCategory,
  priority: number,
  belongsToAllNetworks: boolean,
  countSource: Indirect, // if it is possible for the API to find several identical results of this type and count them, then this field tells us what field in the response contains the count (it can be an array, in this case we will read the length property)
  queryParamField : Indirect, // name of the field in singleAPIresult whose data identifies precisely a result in the back-end
  howToFillresultSuggestionOutput : HowToFillresultSuggestionOutput // will be used at execution time to know what data we must copy into each ResultSuggestion.output
}

export const TypeInfo: Record<ResultType, TypeInfoFields> = {
  [ResultType.Tokens]: {
    title: ['common.erc20token', SINGULAR],
    category: Category.Tokens,
    subCategory: SubCategory.Tokens,
    priority: 3,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIstr_value, // this tells us that field `str_value` in singleAPIresult identifies precisely a result of type ResultType.Tokens when communicating about it with the back-end
    howToFillresultSuggestionOutput: { name: Indirect.APIstr_value, description: '', lowLevelData: Indirect.APIhash_value } // this tells us that field `name` in ResultSuggestionOutput will be filled with the content of `str_value` in singleAPIresult, and `lowLevelData` will be filled with `hash_value`
  },
  [ResultType.NFTs]: {
    title: ['common.nft_as_token', SINGULAR],
    category: Category.NFTs,
    subCategory: SubCategory.NFTs,
    priority: 4,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.APIstr_value, description: '', lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.Epochs]: {
    title: ['common.epoch', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.Epochs,
    priority: 12,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: '' }
  },
  [ResultType.Slots]: {
    title: ['common.slot', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 11,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.Blocks]: {
    title: ['common.block', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 10,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.BlockRoots]: {
    title: ['common.block_root', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 18,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.StateRoots]: {
    title: ['common.state_root', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.SlotsAndBlocks,
    priority: 19,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.Transactions]: {
    title: ['common.transaction', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.Transactions,
    priority: 17,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: '', lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.TransactionBatches]: {
    title: ['common.tx_batch', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 14,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: '' }
  },
  [ResultType.StateBatches]: {
    title: ['common.state_batch', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.Batches,
    priority: 13,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: Indirect.APInum_value, lowLevelData: '' }
  },
  [ResultType.Contracts]: {
    title: ['common.contract', SINGULAR],
    category: Category.Addresses,
    subCategory: SubCategory.Contracts,
    priority: 2,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.APIstr_value, description: '', lowLevelData: Indirect.APIhash_value } // str_value is the name of the contract (for ex: "uniswap") but if the API gives '' we will replace it with a generic name (the title of this type: "Contract")
  },
  [ResultType.Accounts]: {
    title: ['common.account', SINGULAR],
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 2,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: '', lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.EnsAddresses]: {
    title: ['common.ens_address', SINGULAR],
    category: Category.Addresses,
    subCategory: SubCategory.Accounts,
    priority: 1,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.APIstr_value, description: '', lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.EnsOverview]: {
    title: ['common.overview_of_ens', SINGULAR],
    category: Category.Addresses,
    subCategory: SubCategory.EnsOverview,
    priority: 15,
    belongsToAllNetworks: true,
    countSource: Indirect.None,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: Indirect.APIstr_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.Graffiti]: {
    title: ['common.graffiti', SINGULAR],
    category: Category.Protocol,
    subCategory: SubCategory.Graffiti,
    priority: 16,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.TypeTitle, description: ['search_bar.blocks_with', 0], lowLevelData: Indirect.APIstr_value }
  },
  [ResultType.ValidatorsByIndex]: {
    title: ['search_bar.validator_by_index', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APInum_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.ValidatorsByPubkey]: {
    title: ['search_bar.validator_by_public_key', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9,
    belongsToAllNetworks: false,
    countSource: Indirect.None,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: Indirect.APInum_value, lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.ValidatorsByDepositAddress]: {
    title: ['search_bar.validator_by_deposit_address', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 6,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.deposited_by', 0], lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.ValidatorsByDepositEnsName]: {
    title: ['search_bar.validator_by_deposit_ens', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 5,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.deposited_by', 0], lowLevelData: Indirect.APIstr_value }
  },
  [ResultType.ValidatorsByWithdrawalCredential]: {
    title: ['search_bar.validator_by_credential', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.credential', SINGULAR], lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.ValidatorsByWithdrawalAddress]: {
    title: ['search_bar.validator_by_withdrawal_address', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 8,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIhash_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.withdrawn_to', 0], lowLevelData: Indirect.APIhash_value }
  },
  [ResultType.ValidatorsByWithdrawalEnsName]: {
    title: ['search_bar.validator_by_withdrawal_ens', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 7,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.withdrawn_to', 0], lowLevelData: Indirect.APIstr_value }
  },
  [ResultType.ValidatorsByGraffiti]: {
    title: ['search_bar.validator_by_graffiti', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.block_graffiti', 0], lowLevelData: Indirect.APIstr_value }
  }
  /* [ResultType.ValidatorsByName]: {
    title: ['search_bar.validator_by_name', 0],
    category: Category.Validators,
    subCategory: SubCategory.Validators,
    priority: 9999,
    belongsToAllNetworks: false,
    countSource: Indirect.APInum_value,
    queryParamField: Indirect.APIstr_value,
    howToFillresultSuggestionOutput: { name: Indirect.SubCategoryTitle, description: ['search_bar.named', 0], lowLevelData: Indirect.APIstr_value }
  } */
}

export interface PremiumRowCallBackFunction { (result: ResultSuggestion) : boolean }

export interface ExposedSearchbarMethods { // for internal use
  hideResult : (whichOne : ResultSuggestion) => void,
  closeDropdown : () => void
  empty : () => void
}
export interface SearchBar // your ref to the search-bar component must be of this type
       extends ComponentPublicInstance, ExposedSearchbarMethods {}

export type CategoryFilter = Map<Category, boolean> // for internal use
export type NetworkFilter = Map<ChainIDs, boolean> // for internal use
export type SearchbarDropdownLayout = 'narrow-dropdown' | 'large-dropdown' // for internal use (do not change the litterals, they are used as class names)

export function wasOutputDataGivenByTheAPI (type : ResultType, resultSuggestionOutputField : keyof HowToFillresultSuggestionOutput) : boolean {
  switch (TypeInfo[type].howToFillresultSuggestionOutput[resultSuggestionOutputField]) {
    case Indirect.APIstr_value :
    case Indirect.APInum_value :
    case Indirect.APIhash_value :
      return true
    default:
      return false
  }
}

/**
 * @returns Data read from the API response. `undefined` means that something is wrong.
 */
export function realizeData (apiResponseElement: SingleAPIresult, dataSource: FillFrom, t: ComposerTranslation) : any {
  const type = apiResponseElement.type as ResultType
  let sourceField : keyof SingleAPIresult

  switch (dataSource) {
    case Indirect.None : return ''
    case Indirect.APIstr_value : sourceField = 'str_value'; break
    case Indirect.APInum_value : sourceField = 'num_value'; break
    case Indirect.APIhash_value : sourceField = 'hash_value'; break
    case Indirect.CategoryTitle : return t(...CategoryInfo[TypeInfo[type].category].title)
    case Indirect.SubCategoryTitle : return t(...SubCategoryInfo[TypeInfo[type].subCategory].title)
    case Indirect.TypeTitle : return t(...TypeInfo[type].title)
    default :
      return (dataSource === '') ? '' : t(...dataSource)
  }

  return apiResponseElement[sourceField]
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

const searchableTypesPerCategory : Record<Category, ResultType[]> = {} as Record<Category, ResultType[]>
/**
 * @returns the list of types belonging to the given category.
 * This function is fast on average: it computes the lists only at the first call. Subsequent calls return the already computed lists.
 */
export function getListOfResultTypesInCategory (category: Category) : ResultType[] {
  if (!(category in searchableTypesPerCategory)) {
    for (const t of getListOfResultTypes(true)) {
      const c = TypeInfo[t].category
      if (!searchableTypesPerCategory[c]) {
        searchableTypesPerCategory[c] = []
      }
      searchableTypesPerCategory[c].push(t)
    }
  }
  return searchableTypesPerCategory[category]
}

/**
 * This is an example of function that `<BcSearchbarMain>` needs in its props `pick-by-default`. You can design a function fulfilling your needs
 * or simply give this one (after importing `pickHighestPriorityAmongBestMatchings` at the top of your script setup) if it does what you
 * need.
 * What we implemented in this example function:
 * We look for the matching that matches the best with the user input (this is known through the field `Matching.closeness`).
 * If several matchings with this best closeness value exist, we catch the first one (so the one having the highest priority). This
 * happens for example when the user input corresponds to both a validator index and a block number, or both a graffiti and a token name, etc.
 * @param possibilities here the function receives the list of matchings (representing result suggestions)
 * @returns the matching fulfilling the criteria explained above
 */
export function pickHighestPriorityAmongBestMatchings (possibilities : Matching[]) : Matching|undefined {
  let bestMatchWithHigherPriority = possibilities[0]
  for (const possibility of possibilities) {
    if (possibility.closeness < bestMatchWithHigherPriority.closeness) {
      bestMatchWithHigherPriority = possibility
    }
  }
  return bestMatchWithHigherPriority
}

/**
 * @returns the I18n path of a TranslatableLitteral that you can give to t(). Useful to display the litteral in singular or in plural with respect to your needs.
 */
export function getI18nPathOfTranslatableLitteral (litteral: TranslatableLitteral) : string {
  return litteral[0] as string
}
