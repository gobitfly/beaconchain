<script setup lang="ts">
/*
Usage
-----

In your <script setup>, write:
  const mySearchbar = ref<SearchBar>()

In your template, write:
  <BcSearchbarMain
    ref="mySearchbar"
    :bar-style="SearchbarStyle.<what you want to see>"
    :bar-purpose="SearchbarPurpose.<what you need to do>"
    :pick-by-default="<function that picks a result on behalf of the user when they press Enter or the search-button>"
    @go="<your function doing something when a result is selected>"
  />

There are more props that you can give to configure the search bar:
  :only-networks="[<list of chain IDs that the bar is authorized to search over>]" // Without this props, the bar searches over all networks.
  :keep-dropdown-open="true" // When the user selects a result, the drop-down does not close.

The list of possible values for :bar-style is in enum `SearchbarStyle` in file `searchbar.ts`.
The list of possible values for :bar-purpose is in enum `SearchbarPurpose` in file `searchbar.ts`.
You can write your own function for `:pick-by-default`, or give the example function written in `searchbar.ts` if it suits your needs:
   :pick-by-default="pickHighestPriorityAmongBestMatchings"

The handler that you give to props `@go` receives one parameter:
  function myHandler (result : ResultSuggestion)
To get a description of the information carried in the parameter, look at the comments in type `ResultSuggestion` in `searchbar.ts`.

The search-bar offers methods that you can call for further tailoring:
  mySearchbar.value!.hideResult(whichOne : ResultSuggestion) // Removes a result from the drop-down. The result is one of those that you obtained in your `@go` handler.
  mySearchbar.value!.closeDropdown() // Useful when you gave `:keep-dropdown-open="true"` in the props but you still need to close the drop-down in certain cases.
  mySearchbar.value!.empty() // By itself, the search-bar never empties its input field nor its drop-down. You can still clear the search-bar with this method if you want the user to retype from scratch.

Changing the behavior of the search bar thanks to `searchbar.ts`
----------------------------------------------------------------
`searchbar.ts` has been designed as a "configuration file" for the search bar. For example, if the protocol to communicate with the API changes, it might
not be necessary to modify the code of the search-bar. It was developped to be as configurable as possible:

If the API returns a new type of result:
  1. Add this type into the `ResultType` enum of `searchbar.ts`.
  2. Tell the bar how to read/display it by adding a new entry in the `TypeInfo` record.

If the API gets the ability to return a new field in some or all elements of its response array:
  1. Write the name of the new field in `SingleAPIresult`.
  2. Create a reference to it in `Indirect`.
  3. In `TypeInfo`, tell the bar when/where this field must be read (by giving its `Indirect` reference).
  4. Add a case for the reference in function `wasOutputDataGivenByTheAPI()`

If for some type of result you want to change the information / order of the information that the user sees in the corresponding rows of the result-suggestion list:
  1. Locate this result type in record `TypeInfo`.
  2. In that entry, change / swap the references that are in field `howToFillresultSuggestionOutput`.

If you want to change in depth the whole result-suggestion list (to change how every row displays the information):
  1. Add a display-mode in the `SuggestionrowCells` enum in `searchbar.ts`
  2. Update the `SearchbarPurposeInfo` record there to tell the bar which Purpose must use your new mode.
  3. Implement this mode in a new root `<div>` at the end of the `<template>` of `SuggestionRow.vue`.

You can create a new `:bar-purpose` if needed:
  1. Add a purpose name into the `SearchbarPurpose` enum of `searchbar.ts`.
  2. Define this purpose into the `SearchbarPurposeInfo` record.

If you want to add or remove a filter button:
  A. Either you simply need create or modify a purpose to see more/less filters (see above).
  B. Or:
    1. Add/remove an entry in enum `Category`.
    2. Add/remove the corresponding category-title and button-label in enum `CategoryInfo`.
    3. You might need to add/remove an entry in `SubCategoryInfo`.
    4. Update (add/remove/change) all relevant entries in record `TypeInfo` to take properly into account your new categorization of the result types.
    5. Update the entries of `SearchbarPurposeInfo` to take into account the new/removed category.

If you want to change the order of the results in the drop-down, it is a bit less straightforward:
  - To change the order of the results inside each network set, modify the values of fields `priority` in the `TypeInfo` record of `searchbar.ts`.
  - Changing the order of the networks sets is a different task. You would need to change the `priority` fields in the `ChainInfo` record of `networks.ts`.
  Note that if different types or networks have the same priority, two results of these types/networks will appear in the drop-down in the order of their `closeness` values
  (a measure of similarity to the user input).
*/
import { warn } from 'vue'
import { levenshteinDistance } from '~/utils/misc'
import {
  Category,
  ResultType,
  type FillFrom,
  type HowToFillresultSuggestionOutput,
  type ResultSuggestionOutput,
  CategoryInfo,
  SubCategoryInfo,
  TypeInfo,
  getListOfResultTypes,
  wasOutputDataGivenByTheAPI,
  type SingleAPIresult,
  type SearchAheadAPIresponse,
  type ResultSuggestion,
  type OrganizedResults,
  Indirect,
  SearchbarStyle,
  SearchbarPurpose,
  SearchbarPurposeInfo,
  type Matching,
  type PickingCallBackFunction,
  type ExposedSearchbarMethods
} from '~/types/searchbar'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const MinimumTimeBetweenAPIcalls = 1000 // ms

const { t } = useI18n()

const { fetch } = useCustomFetch()
const props = defineProps<{
  barStyle: SearchbarStyle, // look of the bar
  barPurpose: SearchbarPurpose, // what the bar will be used for
  onlyNetworks?: ChainIDs[], // the bar will search on these networks only
  pickByDefault: PickingCallBackFunction // see the declaration of the type to get an explanation
  keepDropdownOpen? : boolean // set to `true` if you want the drop down to stay open when the user clicks a suggestion. You can still close it by calling `<searchbar ref>.value.closeDropdown()` method.
}>()
const emit = defineEmits<{(e: 'go', result : ResultSuggestion) : any}>()

defineExpose<ExposedSearchbarMethods>({ hideResult, closeDropdown, empty })

enum States {
  InputIsEmpty,
  WaitingForResults,
  ApiHasResponded,
  Error,
  UpdateIncoming
}

interface GlobalState {
  state : States,
  functionToCallAfterResultsGetOrganized: Function | null
  showDropdown: boolean
}

let searchableTypes : ResultType[] = []
let allTypesBelongToAllNetworks = false

const debouncer = useDebounceValue<string>('', MinimumTimeBetweenAPIcalls)
watch(debouncer.value, callAPIthenOrganizeResultsThenCallBack)

const inputted = ref('')
let lastKnownInput = ''
const globalState = ref<GlobalState>({
  state: States.InputIsEmpty,
  functionToCallAfterResultsGetOrganized: null,
  showDropdown: false
})

const dropdown = ref<HTMLDivElement>()
const inputFieldAndButton = ref<HTMLDivElement>()
const inputField = ref<HTMLInputElement>()

const userFilters = {
  networks: new Map<ChainIDs, boolean>(), // each entry will have a chain ID as key and the state of the option as value
  categories: new Map<Category, boolean>() // each entry will have a Category as key and the state of the button as value
}

const results = {
  raw: { data: [] } as SearchAheadAPIresponse, // response of the API, without structure nor order
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered-in results, organized
    howManyResultsIn: 0,
    out: { networks: [] } as OrganizedResults, // filtered-out results, organized
    howManyResultsOut: 0
  }
}

function hideResult (whichOne : ResultSuggestion) {
  if (!results.raw.data) {
    return
  }
  const toBeRemoved = results.raw.data.indexOf(whichOne.rawResult)
  if (toBeRemoved >= 0) {
    results.raw.data.splice(toBeRemoved, 1)
    refreshOutputArea()
  }
}

function closeDropdown () {
  globalState.value.showDropdown = false
  inputField.value?.blur()
}

function empty () {
  lastKnownInput = ''
  inputted.value = ''
  resetGlobalState(States.InputIsEmpty)
  results.raw = {}
  clearOrganizedResults()
}

function clearOrganizedResults () {
  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }
  results.organized.howManyResultsIn = 0
  results.organized.howManyResultsOut = 0
}

/**
 *
 * @param state the new state that the search-bar enters
 * @returns old state, so you can read it after the call if you need
 */
function resetGlobalState (state : States) : GlobalState {
  const previousState = { ...globalState.value }

  globalState.value.functionToCallAfterResultsGetOrganized = null
  updateGlobalState(state)

  return previousState
}

function updateGlobalState (state : States) {
  if (state === globalState.value.state && state !== States.UpdateIncoming) {
    // we make sure that Vue re-renders the drop-down although the state does not change
    globalState.value.state = States.UpdateIncoming
    nextTick(() => updateGlobalState(state))
  } else {
    globalState.value.state = state
  }
}

onMounted(() => {
  searchableTypes = []
  allTypesBelongToAllNetworks = true
  // builds the list of all search types that the bar will consider, from the list of searchable categories (obtained through props.barPurpose)
  for (const t of getListOfResultTypes(false)) {
    if (SearchbarPurposeInfo[props.barPurpose].searchable.includes(TypeInfo[t].category) && !SearchbarPurposeInfo[props.barPurpose].unsearchable.includes(t)) {
      searchableTypes.push(t)
      allTypesBelongToAllNetworks &&= TypeInfo[t].belongsToAllNetworks // this variable will be used to know whether it is useless to show the network-filter selector
    }
  }
  // creates the entries storing the state of the category-filter buttons, and deselect them
  for (const s of SearchbarPurposeInfo[props.barPurpose].searchable) {
    userFilters.categories.set(s, false)
  }
  // creates the entries storing the state of the network drop-down, and deselect all networks
  const networks = (props.onlyNetworks !== undefined && props.onlyNetworks.length > 0) ? props.onlyNetworks : getListOfImplementedChainIDs(true)
  for (const nw of networks) {
    userFilters.networks.set(nw, false)
  }
  // listens to clicks outside the component
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
  empty()
})

// closes the drop-down if the user interacts with another part of the page
function listenToClicks (event : Event) {
  if (!dropdown.value || !inputFieldAndButton.value ||
      dropdown.value.contains(event.target as Node) || inputFieldAndButton.value.contains(event.target as Node)) {
    return
  }
  closeDropdown()
}

function inputMightHaveChanged () {
  if (inputted.value === lastKnownInput) {
    return
  }
  if (inputted.value.length === 0) {
    empty()
  } else {
    updateGlobalState(States.WaitingForResults)
    debouncer.bounce(inputted.value, false, true)
    // the debouncer will run callAPIthenOrganizeResultsThenCallBack()
  }
  lastKnownInput = inputted.value
}

function handleKeyPressInInputField (key : string) {
  switch (key) {
    case 'Enter' :
      userPressedSearchButtonOrEnter()
      break
    case 'Escape' :
      closeDropdown()
      break
    default:
      inputMightHaveChanged()
      break
  }
}

function userPressedSearchButtonOrEnter () {
  globalState.value.functionToCallAfterResultsGetOrganized = null
  switch (globalState.value.state) {
    case States.InputIsEmpty : // the user enjoys the sound of clicks
      return
    case States.Error : // the previous API call failed and the user tries again with Enter or with the search button
      updateGlobalState(States.WaitingForResults)
      callAPIthenOrganizeResultsThenCallBack() // we start a new search
      return
    case States.WaitingForResults : // the user pressed Enter or clicked the search button, but the results are not here yet
      globalState.value.functionToCallAfterResultsGetOrganized = userPressedSearchButtonOrEnter // we request to be called again once the communication with the API is complete
      return // in the meantime, we do not proceed further
  }
  // from here, we know that the user pressed Enter or clicked the search button to be redirected by us to the most relevant page

  if (results.organized.howManyResultsIn === 0 && !areThereResultsHiddenByUser()) {
    // nothing matching the input has been found
    return
  }
  // the priority is given to filtered-in results
  let toConsider : OrganizedResults
  if (results.organized.howManyResultsIn > 0) {
    toConsider = results.organized.in
  } else {
    // we default to the filtered-out results if there are results but the drop down does not show them
    toConsider = results.organized.out
  }
  // Builds the list of matchings that the parent component will need when picking one by default (in callback function `props.pickByDefault()`).
  // We guarantee props.pickByDefault() that the list is ordered by network and type priority (the sorting is done in `filterAndOrganizeResults()`).
  const possibilities : Matching[] = []
  for (const network of toConsider.networks) {
    for (const type of network.types) {
      // here we assume that the result with the best `closeness` value is the first one is array `type.suggestions` (see the sorting done in `filterAndOrganizeResults()`)
      possibilities.push({ closeness: type.suggestions[0].closeness, network: network.chainId, type: type.type, s: type.suggestions[0] } as Matching)
    }
  }
  // calling back parent's function in charge of making a choice
  const picked = props.pickByDefault(possibilities)
  if (picked) {
    if (!props.keepDropdownOpen) {
      closeDropdown()
    }
    // calling back parent's function taking action with the result
    emit('go', (picked as any).s as ResultSuggestion)
  }
}

function userClickedSuggestion (suggestion : ResultSuggestion) {
  // calls back parent's function taking action with the result
  if (!props.keepDropdownOpen) {
    closeDropdown()
  }
  emit('go', suggestion)
}

function refreshOutputArea () {
  // updates the result lists with the latest API response and user filters
  filterAndOrganizeResults()
  // refreshes the output area in the drop-down
  updateGlobalState(globalState.value.state)
}

async function callAPIthenOrganizeResultsThenCallBack () {
  const inputWhenAPIgotCalled = inputted.value
  let received : SearchAheadAPIresponse | undefined

  try {
    received = await fetch<SearchAheadAPIresponse>(API_PATH.SEARCH, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: {
        input: inputted.value,
        types: searchableTypes,
        count: isResultCountable(undefined)
      }
    })
  } catch (error) {
    received = undefined
  }
  if (inputted.value !== inputWhenAPIgotCalled) { // result/error outdated. Important: errors are ignored because based on an outdated input.
    return
  }
  if (!received || received.error !== undefined || received.data === undefined) {
    resetGlobalState(States.Error) // the user will see an error message
    return
  }

  results.raw = received
  filterAndOrganizeResults()
  const previousState = resetGlobalState(States.ApiHasResponded)

  previousState.functionToCallAfterResultsGetOrganized?.()
}

// Fills `results.organized` by categorizing, filtering and sorting the data of the API.
function filterAndOrganizeResults () {
  clearOrganizedResults()

  if (results.raw.data === undefined) {
    return
  }

  // determining whether filters are used
  let noNetworkIsSelected = true
  for (const nw of userFilters.networks) {
    noNetworkIsSelected &&= !nw[1]
  }
  let noCategoryIsSelected = true
  for (const cat of userFilters.categories) {
    noCategoryIsSelected &&= !cat[1]
  }

  const resultsIn : ResultSuggestion[] = []
  const resultsOut : ResultSuggestion[] = []
  // filling those two lists
  for (const finding of results.raw.data) {
    const toBeAdded = convertSingleAPIresultIntoResultSuggestion(finding)
    if (!toBeAdded) {
      continue
    }
    // discarding findings that our configuration (given in the props) forbids
    const category = TypeInfo[toBeAdded.type].category
    if ((toBeAdded.chainId !== ChainIDs.Any && !userFilters.networks.has(toBeAdded.chainId)) || !userFilters.categories.has(category)) {
      continue
    }
    // determining whether the finding is filtered in or out, sending it to the corresponding list
    const acceptTheChainID = userFilters.networks.get(toBeAdded.chainId) || noNetworkIsSelected || toBeAdded.chainId === ChainIDs.Any
    const acceptTheCategory = userFilters.categories.get(category) || noCategoryIsSelected
    if (acceptTheChainID && acceptTheCategory) {
      resultsIn.push(toBeAdded)
    } else {
      resultsOut.push(toBeAdded)
    }
  }

  sortResults(resultsIn)
  sortResults(resultsOut)
  fillOrganizedResults(resultsIn, results.organized.in)
  fillOrganizedResults(resultsOut, results.organized.out)
  results.organized.howManyResultsIn = resultsIn.length
  results.organized.howManyResultsOut = resultsOut.length

  // This sorting orders the list of results in the drop down and is fundamental for userPressedSearchButtonOrEnter() as well as props.pickByDefault().
  // Do not alter this sorting without considering the needs of those functions and updating the comments guiding the developpers using the search-bar.
  function sortResults (list : ResultSuggestion[]) {
    list.sort((a, b) => ChainInfo[a.chainId].priority - ChainInfo[b.chainId].priority || TypeInfo[a.type].priority - TypeInfo[b.type].priority || a.closeness - b.closeness)
  }

  function fillOrganizedResults (linearSource : ResultSuggestion[], organizedDestination : OrganizedResults) {
    for (const toBeAdded of linearSource) {
      // Picking from the organized results the network that the finding belongs to. Creates the network if needed.
      let existingNetwork = organizedDestination.networks.findIndex(nwElem => nwElem.chainId === toBeAdded.chainId)
      if (existingNetwork < 0) {
        existingNetwork = -1 + organizedDestination.networks.push({
          chainId: toBeAdded.chainId,
          types: []
        })
      }
      // Picking from the network the type group that the finding belongs to. Creates the type group if needed.
      let existingType = organizedDestination.networks[existingNetwork].types.findIndex(tyElem => tyElem.type === toBeAdded.type)
      if (existingType < 0) {
        existingType = -1 + organizedDestination.networks[existingNetwork].types.push({
          type: toBeAdded.type,
          suggestions: []
        })
      }
      // now we can insert the finding at the right place in the organized results
      organizedDestination.networks[existingNetwork].types[existingType].suggestions.push(toBeAdded)
    }
  }
}

// This function takes a single result element returned by the API and organizes it into an element simpler to handle by the
// code of the search bar (because it is more... organized).
// If the result JSON from the API is somehow unexpected, the function returns `undefined`.
// The fields that the function reads in the API response as well as the place they are stored in our ResultSuggestion.output
// object are given by the filling information in TypeInfo[<result type>].howToFillresultSuggestionOutput in types/searchbar.ts
function convertSingleAPIresultIntoResultSuggestion (apiResponseElement : SingleAPIresult) : ResultSuggestion | undefined {
  if (!(getListOfResultTypes(false) as string[]).includes(apiResponseElement.type)) {
    warn('The API returned an unexpected type of search-ahead result: ', apiResponseElement.type)
    return undefined
  }

  const type = apiResponseElement.type as ResultType
  let chainId : ChainIDs
  if (TypeInfo[type].belongsToAllNetworks) {
    chainId = ChainIDs.Any
  } else {
    chainId = apiResponseElement.chain_id as ChainIDs
  }

  const howToFillresultSuggestionOutput = TypeInfo[type].howToFillresultSuggestionOutput
  const output = {} as ResultSuggestionOutput

  for (const k in howToFillresultSuggestionOutput) {
    const key = k as keyof HowToFillresultSuggestionOutput
    const data = realizeData(apiResponseElement, howToFillresultSuggestionOutput[key])
    if (data === undefined) {
      warn('The API returned a search-ahead result of type ', type, ' with a missing field.')
      return undefined
    } else {
      output[key] = data
    }
  }

  // Defaulting the name to the result type if the API gave ''
  // This is expected to happen in one case: when the back-end does not know the name of a contract, it returns ''
  let nameWasUnknown = false
  if (output.name === '') {
    output.name = t(...TypeInfo[type].title)
    nameWasUnknown = true
  }

  // retrieving the data that identifies this very result in the back-end (will be important for the callback function `@go`)
  const queryParam = realizeData(apiResponseElement, TypeInfo[type].queryParamField) as string

  // Getting the number of identical results found. If the API did not clarify the number results for a countable type, we give NaN.
  let count = 1
  if (isResultCountable(type)) {
    count = (apiResponseElement.num_value === undefined) ? NaN : apiResponseElement.num_value
  }

  // We calculate how far the user input is from the result suggestion of the API (the API completes/approximates inputs, for example for graffiti and token names).
  // It will be needed later to pick the best result suggestion when the user hits Enter, and also in the drop-down to order the suggestions by relevance when several results exist in a type group
  let closeness = Number.MAX_SAFE_INTEGER
  for (const k in output) {
    const key = k as keyof HowToFillresultSuggestionOutput
    if (wasOutputDataGivenByTheAPI(type, key)) {
      const cl = levenshteinDistance(inputted.value, output[key])
      if (cl < closeness) {
        closeness = cl
      }
    }
  }

  return { output, queryParam, closeness, count, chainId, type, rawResult: apiResponseElement, nameWasUnknown }
}

function realizeData (apiResponseElement : SingleAPIresult, dataSource : FillFrom) : string | undefined {
  const type = apiResponseElement.type as ResultType
  let sourceField : keyof SingleAPIresult

  switch (dataSource) {
    case Indirect.SASRstr_value : sourceField = 'str_value'; break
    case Indirect.SASRnum_value : sourceField = 'num_value'; break
    case Indirect.SASRhash_value : sourceField = 'hash_value'; break
    case Indirect.CategoryTitle : return t(...CategoryInfo[TypeInfo[type].category].title)
    case Indirect.SubCategoryTitle : return t(...SubCategoryInfo[TypeInfo[type].subCategory].title)
    case Indirect.TypeTitle : return t(...TypeInfo[type].title)
    default :
      return (dataSource === '') ? '' : t(...dataSource)
  }

  if (apiResponseElement[sourceField] !== undefined) {
    return String(apiResponseElement[sourceField])
  }

  return undefined
}

function isResultCountable (type : ResultType | undefined) : boolean {
  if (type !== undefined) {
    return TypeInfo[type].countable
  }
  // from here, there is uncertainty but we must simply tell whether counting is possible for some results
  if (props.barPurpose === SearchbarPurpose.GlobalSearch) {
    return false // we do not ask the API to count identical results when the bar is versatile (general bar to search anything on the blockchain)
  }
  for (const type of searchableTypes) {
    if (TypeInfo[type].countable) {
      return true
    }
  }
  return false
}

function mustNetworkFilterBeShown () : boolean {
  return userFilters.networks.size >= 2 && !allTypesBelongToAllNetworks
}

function mustCategoryFiltersBeShown () : boolean {
  return userFilters.categories.size >= 2
}

const classForDropdownOpenedOrClosed = computed(() => globalState.value.showDropdown ? 'dropdown-is-opened' : 'dropdown-is-closed')

const classIfDropdownContainsSomething = computed(() => {
  const dropdownContainsSomething = mustNetworkFilterBeShown() || mustCategoryFiltersBeShown() || globalState.value.state !== States.InputIsEmpty
  return dropdownContainsSomething ? 'dropdown-contains-something' : ''
})

function areThereResultsHiddenByUser () : boolean {
  return results.organized.howManyResultsOut > 0
}

function informationIfNoResult () : string {
  let info = t('search_bar.no_result_matches') + ' '

  if (areThereResultsHiddenByUser()) {
    info += t('search_bar.your_filters')
  } else {
    info += t('search_bar.your_input')
  }

  return info
}

function informationIfHiddenResults () : string {
  let info = String(results.organized.howManyResultsOut) + ' '

  info += (results.organized.howManyResultsOut === 1 ? t('search_bar.one_result_hidden') : t('search_bar.several_results_hidden'))

  if (results.organized.howManyResultsIn !== 0) {
    info = '+' + info + ' ' + t('search_bar.by_your_filters')
  } else {
    info = '(' + info + ')'
  }

  return info
}
</script>

<template>
  <div class="anchor" :class="[barStyle, classForDropdownOpenedOrClosed]">
    <div class="whole-component" :class="[barStyle, classForDropdownOpenedOrClosed]" @keydown="(e) => e.stopImmediatePropagation()">
      <div ref="inputFieldAndButton" class="input-and-button" :class="barStyle">
        <input
          ref="inputField"
          v-model="inputted"
          class="p-inputtext inputfield"
          :class="barStyle"
          type="text"
          :placeholder="t(SearchbarPurposeInfo[barPurpose].placeHolder)"
          @keyup="(e) => handleKeyPressInInputField(e.key)"
          @focus="globalState.showDropdown = true"
        >
        <BcSearchbarButton
          class="searchbutton"
          :class="[barStyle, classForDropdownOpenedOrClosed]"
          :bar-style="barStyle"
          :bar-purpose="barPurpose"
          @click="userPressedSearchButtonOrEnter()"
        />
      </div>
      <div v-if="globalState.showDropdown" ref="dropdown" class="dropdown" :class="[barStyle, classIfDropdownContainsSomething]">
        <div v-if="classIfDropdownContainsSomething" class="separation" :class="barStyle" />
        <div v-if="mustNetworkFilterBeShown() || mustCategoryFiltersBeShown()" class="filter-area">
          <BcSearchbarNetworkSelector
            v-if="mustNetworkFilterBeShown()"
            class="filter-networks"
            :live-state="userFilters.networks"
            :bar-style="barStyle"
            @change="refreshOutputArea"
          />
          <BcSearchbarCategorySelectors
            v-if="mustCategoryFiltersBeShown()"
            class="filter-categories"
            :live-state="userFilters.categories"
            :bar-style="barStyle"
            @change="refreshOutputArea"
          />
        </div>
        <div v-if="globalState.state === States.ApiHasResponded" class="output-area" :class="barStyle">
          <div v-for="(network, k) of results.organized.in.networks" :key="network.chainId" class="network-container" :class="barStyle">
            <div v-for="(typ, j) of network.types" :key="typ.type" class="type-container" :class="barStyle">
              <div v-for="(suggestion, i) of typ.suggestions" :key="suggestion.queryParam" class="suggestionrow-container" :class="barStyle">
                <div v-if="i+j+k > 0" class="separation-between-suggestions" :class="barStyle" />
                <BcSearchbarSuggestionRow
                  :suggestion="suggestion"
                  :bar-style="barStyle"
                  :bar-purpose="barPurpose"
                  @click="(e : Event) => {e.stopPropagation(); userClickedSuggestion(suggestion)}"
                />
              </div>
            </div>
          </div>
          <div v-if="results.organized.howManyResultsIn == 0" class="info center">
            {{ informationIfNoResult() }}
          </div>
          <div v-if="areThereResultsHiddenByUser()" class="info bottom">
            {{ informationIfHiddenResults() }}
          </div>
        </div>
        <div v-else class="output-area" :class="barStyle">
          <div v-if="globalState.state === States.WaitingForResults" class="info center">
            {{ t('search_bar.searching') }}
            <BcLoadingSpinner :loading="true" size="small" alignment="default" />
          </div>
          <div v-else-if="globalState.state === States.Error" class="info center">
            {{ t('search_bar.something_wrong') }}
            <IconErrorFace :inline="true" />
            <br>
            {{ t('search_bar.try_again') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.anchor {
  position: relative;
  display: flex;
  margin: auto;
  &.embedded {
    height: 30px;
    &.dropdown-is-opened {
      @media (max-width: 469.9px) { // narrow window/screen
        position: absolute;
        left: 0px;
        right: 0px;
      }
    }
  }
  &.discreet {
    height: 34px;
    @media (min-width: 600px) { // large screen
      width: 460px;
    }
    @media (max-width: 600px) { // mobile
      width: 380px;
    }
  }
  &.gaudy {
    height: 40px;
    @media (min-width: 600px) { // large screen
      width: 735px;
    }
    @media (max-width: 600px) { // mobile
      width: 380px;
    }
  }
}

.dropdown-is-opened {
  z-index: 256;
}

.whole-component {
  @include main.container;
  position: absolute;
  right: 0px;
  left: 0px;

  &.gaudy,
  &.embedded {
    background-color: var(--searchbar-background-gaudy);
    border: 1px solid var(--input-border-color);
  }
  &.discreet {
    background-color: var(--searchbar-background-discreet);
    border: 1px solid transparent;
    &.dropdown-is-opened {
      border: 1px solid var(--searchbar-background-hover-discreet);
    }
  }

  .input-and-button {
    position: relative;
    left: 0px;
    right: 0px;

    .inputfield {
      display:inline-block;
      position: relative;
      box-sizing: border-box;
      width: 100%;
      border: none;
      box-shadow: none;
      background-color: transparent;
      padding-top: 0px;
      padding-bottom: 0px;

      &.gaudy {
        height: 40px;
        padding-right: 41px;
      }
      &.discreet {
        height: 34px;
        padding-right: 35px;
        color: var(--searchbar-text-discreet);
        ::placeholder {
          color: var(--light-grey-4);
        }
      }
      &.embedded {
        height: 30px;
        padding-right: 31px;
      }

      &:placeholder-shown {
        text-overflow: ellipsis;
      }
    }

    .searchbutton {
      position: absolute;
      &.gaudy {
        right: -1px;
        top: -1px;
        width: 42px;
        height: 42px;
      }
      &.discreet {
        right: 0px;
        top: 0px;
        width: 34px;
        height: 34px;
      }
      &.embedded {
        right: -1px;
        top: -1px;
        width: 32px;
        height: 32px;
      }
    }
  }

  .dropdown {
    position: relative;
    left: 0px;
    right: 0px;
    &.dropdown-contains-something {
      padding-bottom: 4px;
    }

    .separation {
      position: relative;
      margin-left: 8px;
      margin-right: 8px;
      height: 1px;
      margin-bottom: 10px;
      &.gaudy {
        background-color: var(--input-border-color);
      }
      &.discreet {
        background-color: var(--searchbar-background-hover-discreet);
      }
      &.embedded {
        background-color: var(--input-border-color);
      }
    }

    .filter-area {
      display: flex;
      flex-wrap: wrap;

      .filter-networks {
        margin-left: 6px;
      }
      .filter-categories {
        margin-left: 6px;
      }
    }

    .output-area {
      position: relative;
      display: flex;
      flex-direction: column;
      max-height: 270px;
      overflow: auto;
      @include fonts.standard_text;
      &.discreet {
        color: var(--searchbar-text-discreet);
      }

      .network-container {
        position: relative;
        display: flex;
        flex-direction: column;

        .type-container {
          position: relative;
          display: flex;
          flex-direction: column;

          .suggestionrow-container {
            position: relative;

            .separation-between-suggestions {
              position: relative;
              display: none;
              margin-left: 8px;
              margin-right: 8px;
              height: 1px;

              &.embedded {
                @media (max-width: 600px) { // mobile
                  display: block;
                }
                background-color: var(--input-border-color);
              }
            }
          }
        }
      }

      .info {
        position: relative;
        @include fonts.standard_text;
        color: var(--text-color-disabled);
        justify-content: center;
        text-align: center;
        align-items: center;
        &.bottom {
          padding-top: 6px;
          margin-top: auto;
        }
        &.center {
          margin-bottom: auto;
          margin-top: auto;
        }
        padding-left: 6px;
        padding-right: 6px;
      }
    }
  }
}
</style>
