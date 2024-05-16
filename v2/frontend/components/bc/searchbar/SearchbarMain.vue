<script setup lang="ts">
/*
 * If you want to change the behavior of the component or the information it displays, it is possible that you simply need to change a few parameters
 * in searchbar.ts rather than altering the code of the component. The possibilities offered by this configuration file are explanined in readme.md
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
  getListOfResultTypesInCategory,
  wasOutputDataGivenByTheAPI,
  type SingleAPIresult,
  type SearchAheadAPIresponse,
  type ResultSuggestion,
  type ResultSuggestionInternal,
  type OrganizedResults,
  Indirect,
  SearchbarStyle,
  SearchbarPurpose,
  SearchbarPurposeInfo,
  type Matching,
  type PickingCallBackFunction,
  type ExposedSearchbarMethods,
  type CategoryFilter,
  type NetworkFilter
} from '~/types/searchbar'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'
import { API_PATH } from '~/types/customFetch'

const MinimumTimeBetweenAPIcalls = 1000 // ms

const { t } = useI18n()

const { fetch } = useCustomFetch()
const props = defineProps<{
  barStyle: SearchbarStyle, // look of the bar
  barPurpose: SearchbarPurpose, // what the bar will be used for
  onlyNetworks?: ChainIDs[], // the bar will search on these networks only
  pickByDefault: PickingCallBackFunction, // see the declaration of the type to get an explanation
  keepDropdownOpen?: boolean // set to `true` if you want the drop down to stay open when the user clicks a suggestion. You can still close it by calling `<searchbar ref>.value.closeDropdown()` method.
}>()
const emit = defineEmits<{(e: 'go', result : ResultSuggestion) : any}>()

defineExpose<ExposedSearchbarMethods>({ hideResult, closeDropdown, empty })

enum States {
  NoText,
  WaitingForResults,
  ApiHasResponded,
  Error,
  UpdateIncoming
}

interface GlobalState {
  state: States,
  functionToCallAfterResultsGetOrganized: Function | null
  showDropdown: boolean
}

let differentialRequests : boolean
let searchableTypes: ResultType[] = []
let allTypesBelongToAllNetworks = false

const globalState = ref<GlobalState>({
  state: States.NoText,
  functionToCallAfterResultsGetOrganized: null,
  showDropdown: false
})

const dropdown = ref<HTMLDivElement>()
const textFieldAndButton = ref<HTMLDivElement>()
const textField = ref<HTMLInputElement>()

let userInputNonce = 0
const userInputNetworks = ref<NetworkFilter>(new Map<ChainIDs, boolean>()) // each entry will have a chain ID as key and the state of the option as value
const userInputCategories = ref<CategoryFilter>(new Map<Category, boolean>()) // each entry will have a Category as key and the state of the button as value
let userInputNoNetworkIsSelected = true
let userInputNoCategoryIsSelected = true
const userInputText = ref<string>('')
let lastKnownText = ''

const nextSearchScope = {
  networks: new Set<ChainIDs>(),
  categories: new Set<Category>()
}

const debouncer = useDebounceValue<number>(0, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, callAPIthenOrganizeResultsThenCallBack)

const results = {
  raw: {
    stringifyiedList: new Set<string>(), // List of results returned by the API, without structure nor order. The list can be built in serveral steps (for a same text input, if the user selects new filters, the list can augment).
    scopeMatrix: {} as Record<ChainIDs, Record<Category, boolean>> // tells which network × category combinations have been explored to obtain the current list of results (as the user can select/deselect successively filters in any order, the scope is not straightforward)
  },
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered-in results, organized
    howManyResultsIn: 0,
    out: { networks: [] } as OrganizedResults, // filtered-out results, organized
    howManyResultsOut: 0
  }
}

function hideResult (whichOne : ResultSuggestion) {
  results.raw.stringifyiedList.delete((whichOne as ResultSuggestionInternal).stringifyiedRawResult)
  // now we update the list of result suggestions
  refreshOutputArea()
}

function closeDropdown () {
  globalState.value.showDropdown = false
  textField.value?.blur()
}

function empty () {
  lastKnownText = ''
  userInputNonce++
  userInputText.value = ''
  resetGlobalState(States.NoText)
  clearRawResults()
  clearOrganizedResults()
}

function clearRawResults () {
  results.raw.stringifyiedList.clear()
  for (const nw in results.raw.scopeMatrix) {
    const network = nw as unknown as ChainIDs
    for (const cat in results.raw.scopeMatrix[network]) {
      const category = cat as unknown as Category
      results.raw.scopeMatrix[network][category] = false
    }
  }
}

function clearOrganizedResults () {
  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }
  results.organized.howManyResultsIn = 0
  results.organized.howManyResultsOut = 0
}

/**
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

function reconfigureSearchbar () {
  differentialRequests = SearchbarPurposeInfo[props.barPurpose].differentialRequests
  closeDropdown()
  empty()
  // builds the list of all search types that the bar will consider, from the list of searchable categories (obtained through props.barPurpose)
  searchableTypes = generateTypesFromCategories(SearchbarPurposeInfo[props.barPurpose].searchable)
  allTypesBelongToAllNetworks = true
  for (const t of searchableTypes) {
    allTypesBelongToAllNetworks &&= TypeInfo[t].belongsToAllNetworks // this variable will be used to know whether it is useless to show the network-filter selector
  }
  // creates the entries storing the state of the network filter, and deselect all networks
  const networks = (props.onlyNetworks !== undefined && props.onlyNetworks.length > 0) ? props.onlyNetworks : getListOfImplementedChainIDs(true)
  userInputNetworks.value.clear()
  for (const nw of networks) {
    userInputNetworks.value.set(nw, false)
  }
  userInputNoNetworkIsSelected = true
  // creates the entries storing the state of the category filter, and deselect all categories
  userInputCategories.value.clear()
  for (const s of SearchbarPurposeInfo[props.barPurpose].searchable) {
    userInputCategories.value.set(s, false)
  }
  userInputNoCategoryIsSelected = true
  // creates the matrix of filters
  for (const nw of userInputNetworks.value) {
    results.raw.scopeMatrix[nw[0]] = {} as Record<Category, boolean>
    for (const cat of userInputCategories.value) {
      results.raw.scopeMatrix[nw[0]][cat[0]] = false
    }
  }
}

watch(() => props, reconfigureSearchbar, { immediate: true })

onMounted(() => {
  // listens to clicks outside the component
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
  empty()
})

// closes the drop-down if the user interacts with another part of the page
function listenToClicks (event : Event) {
  if (!globalState.value.showDropdown || !dropdown.value || !textFieldAndButton.value ||
      dropdown.value.contains(event.target as Node) || textFieldAndButton.value.contains(event.target as Node)) {
    return
  }
  closeDropdown()
}

function textMightHaveChanged () {
  if (userInputText.value === lastKnownText) {
    return
  }
  userInputNonce++
  if (userInputText.value.length === 0) {
    empty()
  } else {
    resetGlobalState(States.WaitingForResults)
    clearRawResults()
    calculateNextSearchScope()
    debouncer.bounce(userInputNonce, false, true)
    // the debouncer will run callAPIthenOrganizeResultsThenCallBack()
  }
  lastKnownText = userInputText.value
}

function handleKeyPressInTextField (key : string) {
  switch (key) {
    case 'Enter' :
      userPressedSearchButtonOrEnter()
      break
    case 'Escape' :
      closeDropdown()
      break
    default:
      textMightHaveChanged()
      break
  }
}

function userFiltersChanged () {
  userInputNonce++
  // determining whether no filter is selected
  userInputNoNetworkIsSelected = true
  for (const nw of userInputNetworks.value) {
    userInputNoNetworkIsSelected &&= !nw[1]
  }
  userInputNoCategoryIsSelected = true
  for (const cat of userInputCategories.value) {
    userInputNoCategoryIsSelected &&= !cat[1]
  }
  // if the text input is empty, our work is done
  if (userInputText.value.length === 0) {
    return
  }
  // determining which networks and categories need to be searched in
  calculateNextSearchScope()
  // if the scope did not widen, we simply update the list of result suggestions shown to the user
  if ((!differentialRequests || nextSearchScope.networks.size + nextSearchScope.categories.size === 0) && globalState.value.state !== States.Error) {
    refreshOutputArea()
  } else {
    // the scope is larger so a new request will be sent to the API
    resetGlobalState(States.WaitingForResults)
    debouncer.bounce(userInputNonce, false, true)
  }
}

function userPressedSearchButtonOrEnter () {
  globalState.value.functionToCallAfterResultsGetOrganized = null
  switch (globalState.value.state) {
    case States.NoText : // the user enjoys the sound of clicks
      return
    case States.Error : // the previous API call failed and the user tries again with Enter or with the search button
      resetGlobalState(States.WaitingForResults)
      callAPIthenOrganizeResultsThenCallBack(userInputNonce) // we start a new search
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
      // here we assume that the result with the best `closeness` value is the first one in array `type.suggestions` (see the sorting done in `filterAndOrganizeResults()`)
      possibilities.push({ closeness: type.suggestions[0].closeness, network: network.chainId, type: type.type, s: type.suggestions[0] } as Matching)
    }
  }
  // calling back parent's function in charge of making a choice
  const pickedMatching = props.pickByDefault(possibilities)
  if (pickedMatching) {
    if (!props.keepDropdownOpen) {
      closeDropdown()
    }
    // calling back parent's function taking action with the result
    emit('go', (pickedMatching as any).s as ResultSuggestion)
  }
}

function userClickedSuggestion (suggestion : ResultSuggestionInternal) {
  // calls back parent's function taking action with the result
  if (!props.keepDropdownOpen) {
    closeDropdown()
  }
  emit('go', suggestion as ResultSuggestion)
}

function refreshOutputArea () {
  // updates the result lists with the latest API response and user filters
  filterAndOrganizeResults()
  // refreshes the output area in the drop-down
  updateGlobalState(globalState.value.state)
}

/**
 * Calculate two lists (`nextSearchScope.networks` and `nextSearchScope.categories`) telling where we need new results from.
 * For each filter that the user deselects, the scope is not shrinked because we can simply hide the corresponding results in the drop down.
 * For each filter added, the scope is augmented properly ("properly": for example, if a new category is selected, it is not sufficient to add it to the set of categories,
 *  all networks currently selected are also needed and those are not necessarily all the networks in the scope due to the path followed by the user
 *  while clicking the filters).
 */
function calculateNextSearchScope () {
  nextSearchScope.networks.clear()
  nextSearchScope.categories.clear()
  if (differentialRequests) {
    for (const nw of userInputNetworks.value) {
      if (!nw[1] && !userInputNoNetworkIsSelected) { continue }
      for (const cat of userInputCategories.value) {
        if (!cat[1] && !userInputNoCategoryIsSelected) { continue }
        if (results.raw.scopeMatrix[nw[0]][cat[0]]) { continue }
        // The previous lines ensure that this network × category combination is not in the scope already.
        // The next two lines inventor the network and the category and the Set object ensures that they are inventored only once.
        nextSearchScope.networks.add(nw[0])
        nextSearchScope.categories.add(cat[0])
      }
    }
  } else {
    for (const nw of userInputNetworks.value) {
      nextSearchScope.networks.add(nw[0])
    }
    for (const cat of userInputCategories.value) {
      nextSearchScope.categories.add(cat[0])
    }
  }
}

/**
 * Once new results are received and added to `results.raw.stringifyiedList`,
 * this function is called to add the newly selected filters to `results.raw.scopeMatrix`.
 */
function saveNewSearchScope () {
  for (const nw of nextSearchScope.networks) {
    for (const cat of nextSearchScope.categories) {
      results.raw.scopeMatrix[nw][cat] = true
    }
  }
}

async function callAPIthenOrganizeResultsThenCallBack (nonceWhenCalled: number) {
  let received : SearchAheadAPIresponse | undefined

  try {
    const networks = Array.from(nextSearchScope.networks)
    const types = generateTypesFromCategories(nextSearchScope.categories)
    received = await fetch<SearchAheadAPIresponse>(API_PATH.SEARCH, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: {
        input: userInputText.value,
        networks,
        types,
        include_validators: areResultsCountable(types, true)
      }
    })
  } catch (error) {
    received = undefined
  }
  if (userInputNonce !== nonceWhenCalled) { // Result outdated so we ignore it. If there is an error, we ignore it too because it is based on an outdated input.
    return
  }
  if (!received || received.error !== undefined || received.data === undefined) {
    resetGlobalState(States.Error) // the user will see an error message
    return
  }

  for (const res of received.data) {
    results.raw.stringifyiedList.add(JSON.stringify(res)) // the Set object ensures that we do not add duplicate results (which can happen when the new scope has some overlap with the previous one)
  }
  saveNewSearchScope()

  filterAndOrganizeResults()
  const previousState = resetGlobalState(States.ApiHasResponded)

  previousState.functionToCallAfterResultsGetOrganized?.()
}

// Fills `results.organized` by categorizing, filtering and sorting the data of the API.
function filterAndOrganizeResults () {
  clearOrganizedResults()

  const resultsIn : ResultSuggestionInternal[] = []
  const resultsOut : ResultSuggestionInternal[] = []
  // filling those two lists
  for (const finding of results.raw.stringifyiedList) {
    const toBeAdded = convertSingleAPIresultIntoResultSuggestion(finding)
    if (!toBeAdded) {
      continue
    }
    // discarding findings that our configuration (given in the props) forbids
    const category = TypeInfo[toBeAdded.type].category
    if ((toBeAdded.chainId !== ChainIDs.Any && !userInputNetworks.value.has(toBeAdded.chainId)) || !userInputCategories.value.has(category)) {
      continue
    }
    // determining whether the finding is filtered in or out, sending it to the corresponding list
    const acceptTheChainID = userInputNetworks.value.get(toBeAdded.chainId) || userInputNoNetworkIsSelected || toBeAdded.chainId === ChainIDs.Any
    const acceptTheCategory = userInputCategories.value.get(category) || userInputNoCategoryIsSelected
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
  function sortResults (list : ResultSuggestionInternal[]) {
    list.sort((a, b) => ChainInfo[a.chainId].priority - ChainInfo[b.chainId].priority || TypeInfo[a.type].priority - TypeInfo[b.type].priority || a.closeness - b.closeness)
  }

  function fillOrganizedResults (linearSource : ResultSuggestionInternal[], organizedDestination : OrganizedResults) {
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
// The fields that the function reads in the API response as well as the place they are stored in our ResultSuggestionInternal.output
// object are given by the filling information in TypeInfo[<result type>].howToFillresultSuggestionOutput in types/searchbar.ts
function convertSingleAPIresultIntoResultSuggestion (stringifyiedRawResult : string) : ResultSuggestionInternal | undefined {
  const apiResponseElement = JSON.parse(stringifyiedRawResult) as SingleAPIresult
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
  if (areResultsCountable([type], false)) {
    const countSource = apiResponseElement[TypeInfo[type].countSource!]
    if (!countSource) {
      count = NaN
    } else {
      count = (Array.isArray(countSource)) ? countSource.length : Number(countSource)
    }
  }

  // We calculate how far the user text is from the result suggestion of the API (the API completes/approximates terms, for example for graffiti and token names).
  // It will be needed later to pick the best result suggestion when the user hits Enter, and also in the drop-down to order the suggestions by relevance when several results exist in a type group
  let closeness = Number.MAX_SAFE_INTEGER
  for (const k in output) {
    const key = k as keyof HowToFillresultSuggestionOutput
    if (wasOutputDataGivenByTheAPI(type, key)) {
      const cl = levenshteinDistance(userInputText.value, output[key])
      if (cl < closeness) {
        closeness = cl
      }
    }
  }

  return { output, queryParam, closeness, count, chainId, type, rawResult: apiResponseElement, stringifyiedRawResult, nameWasUnknown }
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

function areResultsCountable (types: ResultType[], toTellTheAPI: boolean) : boolean {
  if (SearchbarPurposeInfo[props.barPurpose].askAPItoCountResults || !toTellTheAPI) {
    for (const type of types) {
      if (TypeInfo[type].countSource) {
        return true
      }
    }
  }
  return false
}

function generateTypesFromCategories (categories : Set<Category> | Category[]) : ResultType[] {
  let list : ResultType[] = []

  for (const cat of categories) {
    list = list.concat(getListOfResultTypesInCategory(cat))
  }
  return list.filter(type => !SearchbarPurposeInfo[props.barPurpose].unsearchable.includes(type))
}

function mustNetworkFilterBeShown () : boolean {
  return userInputNetworks.value.size >= 2 && !allTypesBelongToAllNetworks
}

function mustCategoryFiltersBeShown () : boolean {
  return userInputCategories.value.size >= 2
}

const classForDropdownOpenedOrClosed = computed(() => globalState.value.showDropdown ? 'dropdown-is-opened' : 'dropdown-is-closed')

const dropdownContainsSomething = computed(() => mustNetworkFilterBeShown() || mustCategoryFiltersBeShown() || globalState.value.state !== States.NoText)

function areThereResultsHiddenByUser () : boolean {
  return !differentialRequests && results.organized.howManyResultsOut > 0
}

function informationIfNoResult () : string {
  let info = t('search_bar.no_result_matches') + ' '

  if (differentialRequests) {
    info += t('search_bar.your_filters') + t('search_bar.or') + t('search_bar.your_input')
  } else if (areThereResultsHiddenByUser()) {
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
      <div ref="textFieldAndButton" class="text-and-button" :class="barStyle">
        <input
          ref="textField"
          v-model="userInputText"
          class="p-inputtext text-field"
          :class="barStyle"
          type="text"
          :placeholder="t(SearchbarPurposeInfo[barPurpose].placeHolder)"
          @keyup="(e) => handleKeyPressInTextField(e.key)"
          @focus="globalState.showDropdown = true"
        >
        <BcSearchbarButton
          class="search-button"
          :class="[barStyle, classForDropdownOpenedOrClosed]"
          :bar-style="barStyle"
          :bar-purpose="barPurpose"
          @click="userPressedSearchButtonOrEnter()"
        />
      </div>
      <div v-if="globalState.showDropdown" ref="dropdown" class="dropdown" :class="barStyle">
        <div v-if="dropdownContainsSomething" class="separation" :class="barStyle" />
        <div v-if="mustNetworkFilterBeShown() || mustCategoryFiltersBeShown()" class="filter-area">
          <BcSearchbarNetworkSelector
            v-if="mustNetworkFilterBeShown()"
            v-model="userInputNetworks"
            class="filter-networks"
            :bar-style="barStyle"
            @change="userFiltersChanged"
          />
          <BcSearchbarCategorySelectors
            v-if="mustCategoryFiltersBeShown()"
            v-model="userInputCategories"
            class="filter-categories"
            :bar-style="barStyle"
            @change="userFiltersChanged"
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
                  @click="(e : Event) => {e.stopPropagation(); /* stopping propagation prevents a bug when the search bar is asked to remove a result, making it smaller so the click appears to be outside */ userClickedSuggestion(suggestion)}"
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
        <div v-else-if="globalState.state === States.WaitingForResults || globalState.state === States.Error" class="output-area" :class="barStyle">
          <div v-if="globalState.state === States.WaitingForResults" class="info center">
            {{ t('search_bar.searching') }}
            <BcLoadingSpinner :loading="true" size="small" alignment="center" />
          </div>
          <div v-else-if="globalState.state === States.Error" class="info center">
            <span>
              {{ t('search_bar.something_wrong') }}
              <IconErrorFace :inline="true" />
            </span>
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
    height: 28px;
    &.dropdown-is-opened {
      @media (max-width: 510px) { // narrow window/screen
        position: absolute;
        left: 0px;
        right: 0px;
        top: 0px;
      }
    }
  }
  &.discreet {
    height: 34px;
    @media (min-width: 600px) { // large screen
      width: 460px;
    }
    @media (max-width: 599.9px) { // mobile
      width: 380px;
    }
  }
  &.gaudy {
    height: 40px;
    @media (min-width: 600px) { // large screen
      width: 735px;
    }
    @media (max-width: 599.9px) { // mobile
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

  .text-and-button {
    position: relative;
    left: 0px;
    right: 0px;

    .text-field {
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
        height: 28px;
        padding-right: 31px;
      }

      &:placeholder-shown {
        text-overflow: ellipsis;
      }
    }

    .search-button {
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
        border-top-left-radius: 0;
        border-bottom-left-radius: 0;
      }
    }
  }

  .dropdown {
    position: relative;
    left: 0px;
    right: 0px;

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
      position: relative;
      display: block;

      .filter-networks {
        position: relative;
        display: inline-block;
        margin-left: var(--padding-small);
      }
      .filter-categories {
        position: relative;
        display: inline-block;
        margin-left: var(--padding-small);
      }
    }

    .output-area {
      position: relative;
      display: flex;
      flex-direction: column;
      max-height: 270px;
      overflow: auto;
      padding-bottom: 4px;

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
              margin-left: 8px;
              margin-right: 8px;
              height: 1px;
              display: none;
              &.embedded {
                @media (max-width: 599.9px) { // mobile
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
        display: flex;
        flex-direction: column;
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
          height: 60px;
        }
        padding-left: 6px;
        padding-right: 6px;
      }
    }
  }
}
</style>
