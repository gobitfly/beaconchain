<!-- eslint-disable vue/max-len -- TODO:   plz fix this -->
<script setup lang="ts">
/*
 * If you want to change the behavior of the component or the information it displays, it is possible that you simply need to change a few parameters
 * in searchbar.ts rather than altering the code of the component. The possibilities offered by this configuration file are explanined in readme.md
 */
import { warn } from 'vue'
import { levenshteinDistance } from '~/utils/misc'
import {
  type Category,
  type CategoryFilter,
  type ExposedSearchbarMethods,
  getListOfResultTypes,
  getListOfResultTypesInCategory,
  type HowToFillresultSuggestionOutput,
  LayoutThreshold,
  type Matching,
  MinimumTimeBetweenAPIcalls,
  type NetworkFilter,
  type OrganizedResults,
  type PickingCallBackFunction,
  type PremiumRowCallBackFunction,
  realizeData,
  type ResultSuggestion,
  type ResultSuggestionInternal,
  type ResultSuggestionOutput,
  type ResultType,
  type SearchAheadAPIresponse,
  type SearchbarColors,
  type SearchbarDropdownLayout,
  type SearchbarPurpose,
  SearchbarPurposeInfo,
  type SearchbarShape,
  type SearchRequest,
  type SingleAPIresult,
  TypeInfo,
  wasOutputDataGivenByTheAPI,
} from '~/types/searchbar'
import {
  ChainIDs, ChainInfo,
} from '~/types/network'
import { API_PATH } from '~/types/customFetch'

const dropdownLayout = ref<SearchbarDropdownLayout>('narrow-dropdown')

defineExpose<ExposedSearchbarMethods>({
  closeDropdown,
  empty,
  hideResult,
})

const { t } = useTranslation()
const { fetch } = useCustomFetch()
const { availableNetworks } = useNetworkStore()

const props = defineProps<{
  barPurpose: SearchbarPurpose, // what the bar will be used for
  barShape: SearchbarShape, // shape of the bar
  colorTheme: SearchbarColors, // colors of the bar and its dropdown
  keepDropdownOpen?: boolean, // set to `true` if you want the drop down to stay open when the user clicks a suggestion. You can still close it by calling `<searchbar ref>.value.closeDropdown()` method.
  onlyNetworks?: ChainIDs[], // the bar will search on these networks only
  pickByDefault: PickingCallBackFunction, // see the declaration of the type to get an explanation
  rowLacksPremiumSubscription?: PremiumRowCallBackFunction, // the bar calls this function for each row and deactivates the row if it returns `true`
  screenWidthCausingSuddenChange: number, // this information is needed by MiddleEllipsis
}>()
const emit = defineEmits<{ (e: 'go', result: ResultSuggestion): any }>()

enum States {
  NoText,
  WaitingForResults,
  ApiHasResponded,
  Error,
  UpdateIncoming,
}

interface GlobalState {
  functionToCallAfterResultsGetOrganized: (() => void) | null,
  showDropdown: boolean,
  state: States,
}

let differentialRequests: boolean
let searchableTypes: ResultType[]
let allTypesBelongToAllNetworks = false

const globalState = ref<GlobalState>({
  functionToCallAfterResultsGetOrganized: null,
  showDropdown: false,
  state: States.NoText,
})

const wholeComponent = ref<HTMLDivElement>()
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
  categories: new Set<Category>(),
  networks: new Set<ChainIDs>(),
}

const debouncer = useDebounceValue<number>(0, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, callAPIthenOrganizeResultsThenCallBack)

const results = {
  organized: {
    howManyResultsIn: 0,
    howManyResultsOut: 0,
    in: { networks: [] } as OrganizedResults, // filtered-in results, organized
    out: { networks: [] } as OrganizedResults, // filtered-out results, organized
  },
  raw: {
    scopeMatrix: {} as Record<ChainIDs, Record<Category, boolean>>, // tells which network × category combinations have been explored to obtain the current list of results (as the user can select/deselect successively filters in any order, the scope is not straightforward)
    stringifyiedList: new Set<string>(), // List of results returned by the API, without structure nor order. The list can be built in serveral steps (for a same text input, if the user selects new filters, the list can augment).
  },
}

function hideResult(whichOne: ResultSuggestion) {
  results.raw.stringifyiedList.delete(
    (whichOne as ResultSuggestionInternal).stringifyiedRawResult,
  )
  // now we update the list of result suggestions
  refreshOutputArea()
}

function closeDropdown() {
  globalState.value.showDropdown = false
  textField.value?.blur()
}

function empty() {
  lastKnownText = ''
  userInputNonce++
  userInputText.value = ''
  resetGlobalState(States.NoText)
  clearRawResults()
  clearOrganizedResults()
}

function clearRawResults() {
  results.raw.stringifyiedList.clear()
  for (const nw in results.raw.scopeMatrix) {
    const network = nw as unknown as ChainIDs
    for (const cat in results.raw.scopeMatrix[network]) {
      const category = cat as unknown as Category
      results.raw.scopeMatrix[network][category] = false
    }
  }
}

function clearOrganizedResults() {
  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }
  results.organized.howManyResultsIn = 0
  results.organized.howManyResultsOut = 0
}

/**
 * @param state the new state that the search-bar enters
 * @returns old state, so you can read it after the call if you need
 */
function resetGlobalState(state: States): GlobalState {
  const previousState = { ...globalState.value }

  globalState.value.functionToCallAfterResultsGetOrganized = null
  updateGlobalState(state)

  return previousState
}

function updateGlobalState(state: States) {
  if (state === globalState.value.state && state !== States.UpdateIncoming) {
    // we make sure that Vue re-renders the drop-down although the state does not change
    globalState.value.state = States.UpdateIncoming
    nextTick(() => updateGlobalState(state))
  }
  else {
    globalState.value.state = state
  }
}

function reconfigureSearchbar() {
  differentialRequests
    = SearchbarPurposeInfo[props.barPurpose].differentialRequests
  closeDropdown()
  empty()
  // builds the list of all search types that the bar will consider, from the list of searchable categories (obtained through props.barPurpose)
  searchableTypes = generateTypesFromCategories(
    SearchbarPurposeInfo[props.barPurpose].searchable,
  )
  allTypesBelongToAllNetworks = true
  for (const t of searchableTypes) {
    allTypesBelongToAllNetworks &&= TypeInfo[t].belongsToAllNetworks // this variable will be used to know whether it is useless to show the network-filter selector
  }
  // creates the entries storing the state of the network filter, and deselect all networks
  const networks
    = props.onlyNetworks !== undefined && props.onlyNetworks.length > 0
      ? props.onlyNetworks
      : availableNetworks.value
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
watch(availableNetworks, reconfigureSearchbar)

let resizingObserver: ResizeObserver
if (isClientSide) {
  resizingObserver = new ResizeObserver((entries) => {
    const newLayout: SearchbarDropdownLayout
      = entries[0].borderBoxSize[0].inlineSize < LayoutThreshold
        ? 'narrow-dropdown'
        : 'large-dropdown'
    if (newLayout !== dropdownLayout.value) {
      // reassigning 'narrow-dropdown' to 'narrow-dropdown' (for ex) is not guaranteed to preserve the pointer, so this trick makes sure that we do not trigger Vue watchers for nothing (draining the battery and slowing down the UI)
      dropdownLayout.value = newLayout
    }
  })
}

onMounted(() => {
  resizingObserver.observe(wholeComponent.value!)
  // listens to clicks outside the component
  document.addEventListener('click', listenToClicks)
})

onBeforeUnmount(() => {
  resizingObserver.unobserve(wholeComponent.value!)
  document.removeEventListener('click', listenToClicks)
  empty()
})

// closes the drop-down if the user interacts with another part of the page
function listenToClicks(event: Event) {
  if (
    !globalState.value.showDropdown
    || !dropdown.value
    || !textFieldAndButton.value
    || dropdown.value.contains(event.target as Node)
    || textFieldAndButton.value.contains(event.target as Node)
  ) {
    return
  }
  closeDropdown()
}

function textMightHaveChanged() {
  if (userInputText.value === lastKnownText) {
    return
  }
  userInputNonce++
  if (userInputText.value.length === 0) {
    empty()
  }
  else {
    resetGlobalState(States.WaitingForResults)
    clearRawResults()
    calculateNextSearchScope()
    debouncer.bounce(userInputNonce, false, true)
    // the debouncer will run callAPIthenOrganizeResultsThenCallBack()
  }
  lastKnownText = userInputText.value
}

function handleKeyPressInTextField(key: string) {
  switch (key) {
    case 'Enter':
      userPressedSearchButtonOrEnter()
      break
    case 'Escape':
      closeDropdown()
      break
    default:
      textMightHaveChanged()
      break
  }
}

function userFiltersChanged() {
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
  if (
    (!differentialRequests
    || nextSearchScope.networks.size + nextSearchScope.categories.size === 0)
    && globalState.value.state !== States.Error
  ) {
    refreshOutputArea()
  }
  else {
    // the scope is larger so a new request will be sent to the API
    resetGlobalState(States.WaitingForResults)
    debouncer.bounce(userInputNonce, false, true)
  }
}

function userPressedSearchButtonOrEnter() {
  globalState.value.functionToCallAfterResultsGetOrganized = null
  switch (globalState.value.state) {
    case States.NoText: // the user enjoys the sound of clicks
      return
    case States.Error: // the previous API call failed and the user tries again with Enter or with the search button
      resetGlobalState(States.WaitingForResults)
      callAPIthenOrganizeResultsThenCallBack(userInputNonce) // we start a new search
      return
    case States.WaitingForResults: // the user pressed Enter or clicked the search button, but the results are not here yet
      globalState.value.functionToCallAfterResultsGetOrganized
        = userPressedSearchButtonOrEnter // we request to be called again once the communication with the API is complete
      return // in the meantime, we do not proceed further
  }
  // from here, we know that the user pressed Enter or clicked the search button to let us select the most relevant result

  if (
    results.organized.howManyResultsIn === 0
    && !areThereResultsHiddenByUser()
  ) {
    // nothing matching the input has been found
    return
  }
  // the priority is given to filtered-in results
  let toConsider: OrganizedResults
  if (results.organized.howManyResultsIn > 0) {
    toConsider = results.organized.in
  }
  else {
    // we default to the filtered-out results if there are results but the drop down does not show them
    toConsider = results.organized.out
  }
  // Builds the list of matchings that the parent component will need when picking one by default (in callback function `props.pickByDefault()`).
  // We guarantee props.pickByDefault() that the list is ordered by network and type priority (the sorting is done in `filterAndOrganizeResults()`).
  const possibilities: Matching[] = []
  for (const network of toConsider.networks) {
    for (const type of network.types) {
      // here we assume that the results in array `type.suggestions` are sorted by `closeness` values (see the sorting done in `filterAndOrganizeResults()`)
      for (const suggestion of type.suggestions) {
        if (!suggestion.lacksPremiumSubscription) {
          possibilities.push({
            closeness: suggestion.closeness,
            network: network.chainId,
            s: suggestion,
            type: type.type,
          } as Matching)
          break // no need to continue, other results of the same type would be indistinguishable in the code of function `props.pickByDefault()` (called below) : the only difference is that their closeness values are worse
        }
      }
    }
  }
  if (possibilities.length) {
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
}

function userClickedSuggestion(suggestion: ResultSuggestionInternal) {
  // calls back parent's function taking action with the result
  if (!props.keepDropdownOpen) {
    closeDropdown()
  }
  emit('go', suggestion as ResultSuggestion)
}

function refreshOutputArea() {
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
function calculateNextSearchScope() {
  nextSearchScope.networks.clear()
  nextSearchScope.categories.clear()
  if (differentialRequests) {
    for (const nw of userInputNetworks.value) {
      if (!nw[1] && !userInputNoNetworkIsSelected) {
        continue
      }
      for (const cat of userInputCategories.value) {
        if (!cat[1] && !userInputNoCategoryIsSelected) {
          continue
        }
        if (results.raw.scopeMatrix[nw[0]][cat[0]]) {
          continue
        }
        // The previous lines ensure that this network × category combination is not in the scope already.
        // The next two lines inventor the network and the category and the Set object ensures that they are inventored only once.
        nextSearchScope.networks.add(nw[0])
        nextSearchScope.categories.add(cat[0])
      }
    }
  }
  else {
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
function saveNewSearchScope() {
  for (const nw of nextSearchScope.networks) {
    for (const cat of nextSearchScope.categories) {
      results.raw.scopeMatrix[nw][cat] = true
    }
  }
}

async function callAPIthenOrganizeResultsThenCallBack(nonceWhenCalled: number) {
  let received: SearchAheadAPIresponse | undefined

  try {
    const networks = Array.from(nextSearchScope.networks)
    const types = generateTypesFromCategories(nextSearchScope.categories)
    const body: SearchRequest = {
      input: userInputText.value,
      networks,
      types,
    }
    if (areResultsCountable(types, true)) {
      body.count = true
    }
    received = await fetch<SearchAheadAPIresponse>(API_PATH.SEARCH, {
      body,
      headers: { 'Content-Type': 'application/json' },
      method: 'POST',
    })
  }
  catch (error) {
    received = undefined
  }
  if (userInputNonce !== nonceWhenCalled) {
    // Result outdated so we ignore it. If there is an error, we ignore it too because it is based on an outdated input.
    return
  }
  if (
    !received
    || received.error !== undefined
    || received.data === undefined
  ) {
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
function filterAndOrganizeResults() {
  clearOrganizedResults()

  const resultsIn: ResultSuggestionInternal[] = []
  const resultsOut: ResultSuggestionInternal[] = []
  // filling those two lists
  for (const finding of results.raw.stringifyiedList) {
    const toBeAdded = convertSingleAPIresultIntoResultSuggestion(finding)
    if (!toBeAdded) {
      continue
    }
    // discarding findings that our configuration (given in the props) forbids
    const category = TypeInfo[toBeAdded.type].category
    if (
      (toBeAdded.chainId !== ChainIDs.Any
      && !userInputNetworks.value.has(toBeAdded.chainId))
      || !userInputCategories.value.has(category)
    ) {
      continue
    }
    // determining whether the finding is filtered in or out, sending it to the corresponding list
    const acceptTheChainID
      = userInputNetworks.value.get(toBeAdded.chainId)
      || userInputNoNetworkIsSelected
      || toBeAdded.chainId === ChainIDs.Any
    const acceptTheCategory
      = userInputCategories.value.get(category) || userInputNoCategoryIsSelected
    if (acceptTheChainID && acceptTheCategory) {
      resultsIn.push(toBeAdded)
    }
    else {
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
  function sortResults(list: ResultSuggestionInternal[]) {
    list.sort(
      (a, b) =>
        ChainInfo[a.chainId].priority - ChainInfo[b.chainId].priority
        || TypeInfo[a.type].priority - TypeInfo[b.type].priority
        || a.closeness - b.closeness,
    )
  }

  function fillOrganizedResults(
    linearSource: ResultSuggestionInternal[],
    organizedDestination: OrganizedResults,
  ) {
    for (const toBeAdded of linearSource) {
      // Picking from the organized results the network that the finding belongs to. Creates the network if needed.
      let existingNetwork = organizedDestination.networks.findIndex(
        nwElem => nwElem.chainId === toBeAdded.chainId,
      )
      if (existingNetwork < 0) {
        existingNetwork
          = -1
          + organizedDestination.networks.push({
            chainId: toBeAdded.chainId,
            types: [],
          })
      }
      // Picking from the network the type group that the finding belongs to. Creates the type group if needed.
      let existingType = organizedDestination.networks[
        existingNetwork
      ].types.findIndex(tyElem => tyElem.type === toBeAdded.type)
      if (existingType < 0) {
        existingType
          = -1
          + organizedDestination.networks[existingNetwork].types.push({
            suggestions: [],
            type: toBeAdded.type,
          })
      }
      // now we can insert the finding at the right place in the organized results
      organizedDestination.networks[existingNetwork].types[
        existingType
      ].suggestions.push(toBeAdded)
    }
  }
}

// This function takes a single result element returned by the API and organizes it into an element simpler to handle by the
// code of the search bar (because it is more... organized).
// If the result JSON from the API is somehow unexpected, the function returns `undefined`.
// The fields that the function reads in the API response as well as the place they are stored in our ResultSuggestionInternal.output
// object are given by the filling information in TypeInfo[<result type>].howToFillresultSuggestionOutput in types/searchbar.ts
function convertSingleAPIresultIntoResultSuggestion(
  stringifyiedRawResult: string,
): ResultSuggestionInternal | undefined {
  const apiResponseElement = JSON.parse(
    stringifyiedRawResult,
  ) as SingleAPIresult
  if (
    !(getListOfResultTypes(false) as string[]).includes(apiResponseElement.type)
  ) {
    warn(
      'The API returned an unexpected type of search-ahead result: ',
      apiResponseElement.type,
    )
    return undefined
  }

  const type = apiResponseElement.type as ResultType
  let chainId: ChainIDs
  if (TypeInfo[type].belongsToAllNetworks) {
    chainId = ChainIDs.Any
  }
  else {
    chainId = apiResponseElement.chain_id as ChainIDs
  }

  const howToFillresultSuggestionOutput
    = TypeInfo[type].howToFillresultSuggestionOutput
  const output = {} as ResultSuggestionOutput

  for (const k in howToFillresultSuggestionOutput) {
    const key = k as keyof HowToFillresultSuggestionOutput
    const data = realizeData(
      apiResponseElement,
      howToFillresultSuggestionOutput[key],
      t,
    )
    if (data === undefined) {
      warn(
        'The API returned a search-ahead result of type ',
        type,
        ' with a missing field.',
      )
      return undefined
    }
    else {
      output[key] = String(data)
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
  const queryParam = String(
    realizeData(apiResponseElement, TypeInfo[type].queryParamField, t),
  )

  // Getting the number of identical results found. If the API did not clarify the number results for a countable type, we give NaN.
  let count = 1
  if (areResultsCountable([ type ], false)) {
    const countSource = realizeData(
      apiResponseElement,
      TypeInfo[type].countSource,
      t,
    )
    if (countSource === undefined) {
      count = NaN
    }
    else {
      count = Array.isArray(countSource)
        ? countSource.length
        : Number(countSource)
    }
    if (
      (SearchbarPurposeInfo[props.barPurpose].askAPItoCountResults
      && isNaN(count))
      || count <= 0
    ) {
      warn(
        'The API returned a search-ahead result of type ',
        type,
        ' but the batch or count data is missing or wrong.',
      )
      return undefined
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

  const result = {
    chainId,
    closeness,
    count,
    output,
    queryParam,
    rawResult: apiResponseElement,
    type,
  }
  const lacksPremiumSubscription
    = !!props.rowLacksPremiumSubscription
    && props.rowLacksPremiumSubscription(result)

  return {
    ...result,
    lacksPremiumSubscription,
    nameWasUnknown,
    stringifyiedRawResult,
  }
}

function areResultsCountable(
  types: ResultType[],
  toTellTheAPI: boolean,
): boolean {
  if (
    SearchbarPurposeInfo[props.barPurpose].askAPItoCountResults
    || !toTellTheAPI
  ) {
    for (const type of types) {
      if (TypeInfo[type].countSource) {
        return true
      }
    }
  }
  return false
}

function generateTypesFromCategories(
  categories: Category[] | Set<Category>,
): ResultType[] {
  let list: ResultType[] = []

  for (const cat of categories) {
    list = list.concat(getListOfResultTypesInCategory(cat))
  }
  return list.filter(
    type =>
      !SearchbarPurposeInfo[props.barPurpose].unsearchable.includes(type),
  )
}

function mustNetworkFilterBeShown(): boolean {
  return userInputNetworks.value.size >= 2 && !allTypesBelongToAllNetworks
}

function mustCategoryFiltersBeShown(): boolean {
  return userInputCategories.value.size >= 2
}

const classForDropdownOpenedOrClosed = computed(() =>
  globalState.value.showDropdown ? 'dropdown-is-opened' : 'dropdown-is-closed',
)

const dropdownContainsSomething = computed(
  () =>
    mustNetworkFilterBeShown()
    || mustCategoryFiltersBeShown()
    || globalState.value.state !== States.NoText,
)

function areThereResultsHiddenByUser(): boolean {
  return !differentialRequests && results.organized.howManyResultsOut > 0
}

function informationIfNoResult(): string {
  let info = t('search_bar.no_result_matches') + ' '

  if (differentialRequests) {
    info += t('search_bar.your_input')
    if (!userInputNoNetworkIsSelected || !userInputNoCategoryIsSelected) {
      info += t('search_bar.or') + t('search_bar.your_filters')
    }
  }
  else if (areThereResultsHiddenByUser()) {
    info += t('search_bar.your_filters')
  }
  else {
    info += t('search_bar.your_input')
  }
  return info
}

function informationIfHiddenResults(): string {
  let info = String(results.organized.howManyResultsOut) + ' '

  info
    += results.organized.howManyResultsOut === 1
      ? t('search_bar.one_result_hidden')
      : t('search_bar.several_results_hidden')

  if (results.organized.howManyResultsIn !== 0) {
    info = '+' + info + ' ' + t('search_bar.by_your_filters')
  }
  else {
    info = '(' + info + ')'
  }

  return info
}
</script>

<template>
  <div
    class="anchor"
    :class="[barShape, classForDropdownOpenedOrClosed]"
  >
    <div
      ref="wholeComponent"
      class="whole-component"
      :class="[barShape, colorTheme, classForDropdownOpenedOrClosed]"
      @keydown="(e) => e.stopImmediatePropagation()"
    >
      <div
        ref="textFieldAndButton"
        class="text-and-button"
        :class="barShape"
      >
        <input
          ref="textField"
          v-model="userInputText"
          class="p-inputtext text-field"
          :class="[barShape, colorTheme]"
          type="text"
          :placeholder="t(SearchbarPurposeInfo[barPurpose].placeHolder)"
          @keyup="(e) => handleKeyPressInTextField(e.key)"
          @focus="globalState.showDropdown = true"
        >
        <BcSearchbarButton
          class="search-button"
          :class="[barShape, classForDropdownOpenedOrClosed]"
          :bar-shape
          :color-theme
          :bar-purpose
          @click="userPressedSearchButtonOrEnter()"
        />
      </div>
      <div
        v-if="globalState.showDropdown"
        ref="dropdown"
        class="dropdown"
        :class="barShape"
      >
        <div
          v-if="dropdownContainsSomething"
          class="separation"
          :class="[barShape, colorTheme]"
        />
        <div
          v-if="mustNetworkFilterBeShown() || mustCategoryFiltersBeShown()"
          class="filter-area"
        >
          <BcSearchbarNetworkSelector
            v-if="mustNetworkFilterBeShown()"
            v-model="userInputNetworks"
            class="filter-networks"
            :bar-shape
            :color-theme
            :dropdown-layout
            @change="userFiltersChanged"
          />
          <BcSearchbarCategorySelectors
            v-if="mustCategoryFiltersBeShown()"
            v-model="userInputCategories"
            class="filter-categories"
            :bar-shape
            :color-theme
            :dropdown-layout
            @change="userFiltersChanged"
          />
        </div>
        <div
          v-if="globalState.state === States.ApiHasResponded"
          class="output-area"
          :class="[barShape, colorTheme]"
        >
          <div
            v-for="(network, k) of results.organized.in.networks"
            :key="network.chainId"
            class="network-container"
            :class="barShape"
          >
            <div
              v-for="(typ, j) of network.types"
              :key="typ.type"
              class="type-container"
              :class="barShape"
            >
              <div
                v-for="(suggestion, i) of typ.suggestions"
                :key="suggestion.queryParam"
                class="suggestionrow-container"
                :class="barShape"
              >
                <div
                  v-if="i + j + k > 0"
                  class="separation-between-suggestions"
                  :class="[barShape, dropdownLayout]"
                />
                <BcSearchbarSuggestionRow
                  :suggestion
                  :bar-shape
                  :color-theme
                  :dropdown-layout
                  :bar-purpose
                  :screen-width-causing-sudden-change
                  @click="
                    (e: Event) => {
                      e.stopPropagation();
                      /* stopping propagation prevents a bug when the search bar is asked to remove a result,
                      making it smaller so the click appears to be outside */
                      userClickedSuggestion(
                        suggestion,
                      );
                    }
                  "
                />
              </div>
            </div>
          </div>
          <div
            v-if="results.organized.howManyResultsIn == 0"
            class="info center"
          >
            {{ informationIfNoResult() }}
          </div>
          <div
            v-if="areThereResultsHiddenByUser()"
            class="info bottom"
          >
            {{ informationIfHiddenResults() }}
          </div>
        </div>
        <div
          v-else-if="
            globalState.state === States.WaitingForResults
              || globalState.state === States.Error
          "
          class="output-area"
          :class="[barShape, colorTheme]"
        >
          <div
            v-if="globalState.state === States.WaitingForResults"
            class="info center"
          >
            {{ t("search_bar.searching") }}
            <BcLoadingSpinner
              :loading="true"
              size="small"
              alignment="center"
            />
          </div>
          <div
            v-else-if="globalState.state === States.Error"
            class="info center"
          >
            <span>
              {{ t("search_bar.something_wrong") }}
              <IconErrorFace :inline="true" />
            </span>
            {{ t("search_bar.try_again") }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/fonts.scss";

.anchor {
  position: relative;
  display: flex;
  margin: auto;
  align-items: unset !important;
  flex-wrap: wrap !important;
  white-space: normal !important;
  background-color: transparent !important;
  border: none !important;

  &.small {
    height: 28px;
    &.dropdown-is-opened {
      @media (max-width: 510px) {
        // narrow window/screen
        position: absolute;
        left: 0px;
        right: 0px;
        top: 0px;
      }
    }
  }
  &.medium {
    height: 34px;
  }
  &.big {
    height: 40px;
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

  &.default {
    background-color: var(--searchbar-background-default);
    border: 1px solid var(--input-border-color);
  }
  &.darkblue {
    background-color: var(--searchbar-background-darkblue);
    border: 1px solid transparent;
    &.dropdown-is-opened {
      border: 1px solid var(--searchbar-background-hover-darkblue);
    }
  }
  &.lightblue {
    background-color: var(--searchbar-background-lightblue);
    border: 1px solid transparent;
    &.dropdown-is-opened {
      border: 1px solid var(--searchbar-background-hover-lightblue);
    }
  }

  .text-and-button {
    position: relative;
    left: 0px;
    right: 0px;

    .text-field {
      display: inline-block;
      position: relative;
      box-sizing: border-box;
      width: 100%;
      border: none;
      box-shadow: none;
      background-color: transparent;
      padding-top: 0px;
      padding-bottom: 0px;

      &.big {
        height: 40px;
        padding-right: 41px;
      }
      &.medium {
        height: 34px;
        padding-right: 35px;
      }
      &.small {
        height: 28px;
        padding-right: 31px;
      }
      &.darkblue {
        color: var(--searchbar-text-blue);
        &::placeholder {
          color: var(--light-grey);
        }
      }
      &.lightblue {
        color: var(--searchbar-text-blue);
        &::placeholder {
          color: var(--grey-4);
        }
      }

      &:placeholder-shown {
        text-overflow: ellipsis;
      }
    }

    .search-button {
      position: absolute;
      &.big {
        right: -1px;
        top: -1px;
        width: 42px;
        height: 42px;
      }
      &.medium {
        right: 0px;
        top: 0px;
        width: 34px;
        height: 34px;
      }
      &.small {
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
      &.default {
        background-color: var(--input-border-color);
      }
      &.darkblue {
        background-color: var(--searchbar-background-hover-darkblue);
      }
      &.lightblue {
        background-color: var(--searchbar-background-hover-lightblue);
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

      &.darkblue,
      &.lightblue {
        color: var(--searchbar-text-blue);
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
              background-color: var(--input-border-color);
              &.small {
                &.narrow-dropdown {
                  display: block;
                }
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
        padding-left: 6px;
        padding-right: 6px;
        &.bottom {
          padding-top: 6px;
          margin-top: auto;
        }
        &.center {
          margin-bottom: auto;
          margin-top: auto;
          height: 60px;
        }
      }
    }
  }
}
</style>
