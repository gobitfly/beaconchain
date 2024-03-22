<script setup lang="ts">
import { warn } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faMagnifyingGlass } from '@fortawesome/pro-solid-svg-icons'
import {
  Category,
  ResultType,
  TypeInfo,
  getListOfResultTypes,
  type SearchAheadSingleResult,
  type SearchAheadResult,
  type ResultSuggestion,
  type OrganizedResults,
  type SearchBarStyle,
  type Matching,
  type PickingCallBackFunction
} from '~/types/searchbar'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const props = defineProps<{
  searchable: Category[], // list of categories that the bar can search in
  barStyle: SearchBarStyle, // look of the bar ('discreet', 'gaudy' or 'embedded')
  pickByDefault: PickingCallBackFunction // when the user presses Enter, this callback function receives a simplified representation of the possible matches and must return one element from this list. The parameter (of type Matching[]) is a simplified view of the list of results sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. The bar will then trigger the event `@go` to call your handler with the result data of the matching that you picked.
}>()
const emit = defineEmits(['go'])

enum ResultState {
  Obtained, Outdated, Error
}

enum States {
  InputIsEmpty,
  SearchRequestWillBeSent,
  WaitingForResults,
  ApiHasResponded,
  Error,
  UpdateIncoming
}
interface GlobalState {
  state : States,
  callAgainFunctionUserPressedSearchButtonOrEnter: boolean
  showDropDown: boolean
}

const SearchRequestPeriodicity = 2 * 1000 // 2 seconds

const barStyle : SearchBarStyle = props.barStyle as SearchBarStyle
const searchButtonSize = (barStyle === 'discreet') ? '34px' : '40px'

const searchable = props.searchable as Category[]
let searchableTypes : ResultType[] = []

const inputted = ref('')
let lastKnownInput = ''
const globalState = ref<GlobalState>({
  state: States.InputIsEmpty,
  callAgainFunctionUserPressedSearchButtonOrEnter: false,
  showDropDown: false
})
const dropDown = ref<HTMLDivElement>()
const inputFieldAndButton = ref<HTMLDivElement>()

const results = {
  raw: { data: [] } as SearchAheadResult, // response of the API, without structure nor order
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered-in results, organized
    howManyResultsIn: 0,
    out: { networks: [] } as OrganizedResults, // filtered-out results, organized
    howManyResultsOut: 0
  }
}

const userFilters = {
  networks: {} as Record<string, boolean>, // each field will have a String(ChainIDs) as key and the state of the option as value
  noNetworkIsSelected: true,
  categories: {} as Record<string, boolean>, // each field will have a Category as key and the state of the button as value
  noCategoryIsSelected: true
}

function cleanUp (closeDropDown : boolean) {
  lastKnownInput = ''
  inputted.value = ''
  resetGlobalState(States.InputIsEmpty)
  if (closeDropDown) {
    globalState.value.showDropDown = false
  }
}

function resetGlobalState (state : States) : GlobalState {
  const previousState = { ...globalState.value }

  if (state === globalState.value.state) {
    // makes sure that Vue re-renders the drop-down although the state does not change
    globalState.value.state = States.UpdateIncoming
  }
  globalState.value.callAgainFunctionUserPressedSearchButtonOrEnter = false
  globalState.value.state = state

  return previousState
}

function updateGlobalState (state : States) {
  if (state === globalState.value.state) {
    // makes sure that Vue re-renders the drop-down although the state does not change
    globalState.value.state = States.UpdateIncoming
  }
  globalState.value.state = state
}

onMounted(() => {
  searchableTypes = []
  // builds the list of all search types that the bar will consider, from the list of searchable categories (obtained as a props)
  for (const t of getListOfResultTypes(false)) {
    if (searchable.includes(TypeInfo[t].category)) {
      searchableTypes.push(t)
    }
  }
  // creates the fields storing the state of the filter buttons, and deselect them
  for (const s of searchable) {
    userFilters.categories[s] = false
  }
  userFilters.noCategoryIsSelected = true
  // creates the fields storing the state of the network drop-down, and deselect all of them
  for (const nw of getListOfImplementedChainIDs(true)) {
    userFilters.networks[String(nw)] = false
  }
  userFilters.noNetworkIsSelected = true
  // listens to clicks outside the component
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
})

// closes the drop-down if the user interacts with another part of the page
function listenToClicks (event : Event) {
  if (dropDown.value === null || dropDown.value === undefined || inputFieldAndButton.value === undefined ||
      dropDown.value.contains(event.target as Node) || inputFieldAndButton.value.contains(event.target as Node)) {
    return
  }
  globalState.value.showDropDown = false
}

// In the V1, the server was receiving a request between 1.5 and 3.5 seconds after the user inputted something, depending on the length of the input.
// Therefore, the average delay was ~2.5 s for the user as well as for the server. Most of the time the delay was shorter because the 3.5 s delay
// was only for entries of size 1.
// This less-than-2.5s-on-average delay arised from a Timeout Timer.
// For the V2, I propose to work with an Interval Timer because:
// - it makes sure that requests are not sent to the server more often than every 2 s (equivalent to V1),
// - while offering the user an average waiting time of only 1 second through the magic of statistics (better than V1).
setInterval(() => {
  if (globalState.value.state !== States.SearchRequestWillBeSent) {
    return
  }
  updateGlobalState(States.WaitingForResults)

  // These two calls run in a separate thread. They request results from the API and then update the drop-down.
  searchAhead().then(updateBarAfterSearchAhead)
  // the timer returns immediately
},
SearchRequestPeriodicity
)

async function searchAhead () : Promise<ResultState> {
  const startInput = inputted.value
  let received : SearchAheadResult | undefined

  try {
    received = await fetch<SearchAheadResult>(API_PATH.SEARCH, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: { input: inputted.value, searchable: searchableTypes }
    })
  } catch (error) {
    received = undefined
  }
  if (inputted.value !== startInput) { // important: errors are ignored if outdated
    return ResultState.Outdated
  }
  if (received === undefined || received.error !== undefined || received.data === undefined) {
    return ResultState.Error
  }
  results.raw = received
  return ResultState.Obtained
}

function updateBarAfterSearchAhead (howSearchWent : ResultState) {
  switch (howSearchWent) {
    case ResultState.Error :
      resetGlobalState(States.Error) // the user will see an error message
      break
    case ResultState.Outdated :
      // nothing to do
      return
    case ResultState.Obtained :
      filterAndOrganizeResults()
      // we change the state of the component to States.ApiHasResponded and we check whether callAgainFunctionUserPressedSearchButtonOrEnter was true before the change
      if (resetGlobalState(States.ApiHasResponded).callAgainFunctionUserPressedSearchButtonOrEnter) {
      // userPressedSearchButtonOrEnter() asked to be called again because the user pressed Enter or the search button but the results were still pending
        userPressedSearchButtonOrEnter()
      }
      break
  }
}

function userPressedSearchButtonOrEnter () {
  switch (globalState.value.state) {
    case States.InputIsEmpty : // the user enjoys the sounds of clicks
      return
    case States.Error : // the previous API call failed and the user tries again with Enter or with the search button
      resetGlobalState(States.SearchRequestWillBeSent) // we order a new search (the timer will launch it)
      return
    case States.SearchRequestWillBeSent :
    case States.WaitingForResults : // the user pressed Enter or clicked the search button, but the results are not here yet
      globalState.value.callAgainFunctionUserPressedSearchButtonOrEnter = true // we ask the timer to call this function again when the communication with the API is complete
      return // in the meantime, we do not proceed further
  }
  // from here, we know that the user pressed Enter or clicked the search button to be redirected by us to the most relevant page

  if (results.organized.howManyResultsIn + results.organized.howManyResultsOut === 0) {
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
  // Builds the list of matchings that the parent component will need to pick one by default (in callback function `props.pickByDefault()`).
  // We guarantee props.pickByDefault() that the list is ordered by network and type priority (the sorting is done in `filterAnsdOrganizeResults()`).
  const possibilities : Matching[] = []
  for (const network of toConsider.networks) {
    for (const type of network.types) {
      // here we assume that the result with the best `closeness` value is the first one is array `type.suggestion` (see the sorting done in `filterAnsdOrganizeResults()`)
      possibilities.push({ closeness: type.suggestion[0].closeness, network: network.chainId, type: type.type })
    }
  }
  // calling back parent's function in charge of making a choice
  const picked = props.pickByDefault(possibilities)
  // retrieving the result corresponding to the choice
  const network = toConsider.networks.find(nw => nw.chainId === picked.network)
  const type = network?.types.find(ty => ty.type === picked.type)
  // calling back parent's function taking action with the result
  cleanUp(true)
  emit('go', type?.suggestion[0].columns[type?.suggestion[0].queryParam], type?.type, network?.chainId)
}

function userClickedSuggestion (chain : ChainIDs, type : ResultType, wanted: string) {
  // cleans up and calls back parent's function
  cleanUp(true)
  emit('go', wanted, type, chain)
}

function inputMightHaveChanged () {
  if (inputted.value === lastKnownInput) {
    return
  }
  lastKnownInput = inputted.value
  if (inputted.value.length === 0) {
    cleanUp(false)
  } else {
    // we order a search (the timer will launch it)
    resetGlobalState(States.SearchRequestWillBeSent)
  }
}

function networkFilterHasChanged (state : Record<string, boolean>) {
  let noNetworkIsSelected = true

  for (const nw in userFilters.networks) {
    userFilters.networks[nw] = state[nw]
    noNetworkIsSelected &&= !state[nw]
  }
  userFilters.noNetworkIsSelected = noNetworkIsSelected

  refreshOutputArea()
}

function categoryFilterHasChanged (state : Record<string, boolean>) {
  let noCategoryIsSelected = true

  for (const cat in userFilters.categories) {
    userFilters.categories[cat] = state[cat]
    noCategoryIsSelected &&= !state[cat]
  }
  userFilters.noCategoryIsSelected = noCategoryIsSelected

  refreshOutputArea()
}

// Fills `results.organized` by categorizing, filtering and sorting the data of the API.
function filterAndOrganizeResults () {
  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }
  results.organized.howManyResultsIn = 0
  results.organized.howManyResultsOut = 0

  if (results.raw.data === undefined) {
    return
  }

  for (const finding of results.raw.data) {
    const type = finding.type as ResultType

    // getting organized information from the finding
    const toBeAdded = convertSearchAheadResultIntoResultSuggestion(finding)
    if (toBeAdded.columns.length === 0) {
      continue
    }
    // determining the network that the finding belongs to
    let chainId : ChainIDs
    if (TypeInfo[type].belongsToAllNetworks) {
      chainId = ChainIDs.Any
    } else {
      chainId = finding.chain_id as ChainIDs
    }
    // determining whether the finding is filtered in or out, pointing `place` to the corresponding organized object
    let place : OrganizedResults
    if ((userFilters.networks[String(chainId)] || userFilters.noNetworkIsSelected || chainId === ChainIDs.Any) &&
        (userFilters.categories[TypeInfo[type].category] || userFilters.noCategoryIsSelected)) {
      place = results.organized.in
      results.organized.howManyResultsIn++
    } else {
      place = results.organized.out
      results.organized.howManyResultsOut++
    }
    // Picking from the organized results the network that the finding belongs to. Creates the network if needed.
    let existingNetwork = place.networks.findIndex(nwElem => nwElem.chainId === chainId)
    if (existingNetwork < 0) {
      existingNetwork = -1 + place.networks.push({
        chainId,
        types: []
      })
    }
    // Picking from the network the type group that the finding belongs to. Creates the type group if needed.
    let existingType = place.networks[existingNetwork].types.findIndex(tyElem => tyElem.type === type)
    if (existingType < 0) {
      existingType = -1 + place.networks[existingNetwork].types.push({
        type,
        suggestion: []
      })
    }
    // now we can insert the finding at the right place in the organized results
    place.networks[existingNetwork].types[existingType].suggestion.push(toBeAdded)
  }

  // This sorting orders the displayed results and is fundamental for function userPressedSearchButtonOrEnter(). Do not alter the sorting without considering the needs of that function.
  function sortResults (place : OrganizedResults) {
    place.networks.sort((a, b) => ChainInfo[a.chainId].priority - ChainInfo[b.chainId].priority)
    for (const network of place.networks) {
      network.types.sort((a, b) => TypeInfo[a.type].priority - TypeInfo[b.type].priority)
      for (const type of network.types) {
        type.suggestion.sort((a, b) => a.closeness - b.closeness)
      }
    }
  }
  sortResults(results.organized.in)
  sortResults(results.organized.out)
}

// This function takes a single result element returned by the API and organizes it into an element
// simpler to handle by the code of the search bar (not only for displaying).
// If the result element from the API is somehow unexpected, then the function returns an empty array.
// The fields that the function read in the API response as well as the place they are displayed
// in the drop-down are set in the object `TypeInfo` filled in types/searchbar.ts, by its properties
// dataInSearchAheadResult (sets the fields to read and their order) and dropdownColumns (sets the columns to fill with that ordered data).
function convertSearchAheadResultIntoResultSuggestion (apiResponseElement : SearchAheadSingleResult) : ResultSuggestion {
  const emptyResult : ResultSuggestion = { columns: [], queryParam: -1, closeness: NaN }

  if (!(getListOfResultTypes(false) as string[]).includes(apiResponseElement.type)) {
    warn('The API returned an unexpected type of search-ahead result: ', apiResponseElement.type)
    return emptyResult
  }

  const type = apiResponseElement.type as ResultType
  const columns = Array.from(TypeInfo[type].dropdownColumns)
  let queryParam : number = 0

  // Filling the empty columns of the drop down (some are already filled statically by TypeInfo[type].dropdownColumns)
  // We fill them by taking the API data in the order defined in TypeInfo[type].dataInSearchAheadResult
  const fieldsContainingData = TypeInfo[type].dataInSearchAheadResult
  const queryParamField = fieldsContainingData[TypeInfo[type].queryParamIndex]
  for (const field of fieldsContainingData) {
    if (!apiResponseElement[field]) {
      warn('The API returned a search-ahead result of type ', type, ' with a missing field: ', field)
      return emptyResult
    }
    // Searching for the column to fill with the API data (this nested loop might look inefficient but an optimization would be an overkill (our two arrays are of size 3), so would grow the code without effect)
    for (let i = 0; i < columns.length; i++) {
      if (columns[i] === undefined) {
        // The column to fill is found.
        columns[i] = String(apiResponseElement[field])
        if (field === queryParamField) {
          queryParam = i
        }
        break
      }
    }
  }

  if (columns[0] === '') {
    // Defaulting to the name of the result type.
    // This is useful for example with contracts, when the back-end does not know the name of a contract, the first columns shows "Contract"
    columns[0] = TypeInfo[type].title
  }

  // We calculate how far the user input is from the result suggestion of the API (the API completes/approximates inputs, for example for graffiti).
  // It will be needed later to pick the best result suggestion when the user hits Enter, and also in the drop-down to order the suggestions when several results exist in a type group
  let closeness = Number.MAX_SAFE_INTEGER
  for (const field of TypeInfo[type].dataInSearchAheadResult) {
    const cl = resemblanceWithInput(String(apiResponseElement[field]))
    if (cl < closeness) {
      closeness = cl
    }
  }

  return { columns: columns as string[], queryParam, closeness }
}

// Calculates the Levenshtein distance between the parameter and the user input.
// lower value means better similarity and vice-versa
function resemblanceWithInput (str2 : string) : number {
  const str1 = inputted.value
  const dist = []

  for (let i = 0; i <= str1.length; i++) {
    dist[i] = [i]
    for (let j = 1; j <= str2.length; j++) {
      if (i === 0) {
        dist[i][j] = j
      } else {
        const subst = (str1[i - 1] === str2[j - 1]) ? 0 : 1
        dist[i][j] = Math.min(dist[i - 1][j] + 1, dist[i][j - 1] + 1, dist[i - 1][j - 1] + subst)
      }
    }
  }
  return dist[str1.length][str2.length]
}

function refreshOutputArea () {
  // updates the result lists with the latest API response and user filters
  filterAndOrganizeResults()
  // refreshes the output area in the drop-down
  updateGlobalState(globalState.value.state)
}
</script>

<template>
  <div class="anchor" :class="barStyle">
    <div class="whole-component" :class="[barStyle, globalState.showDropDown?'dropdown-is-opened':'']">
      <div ref="inputFieldAndButton" class="input-and-button" :class="barStyle">
        <InputText
          v-model="inputted"
          class="input-field"
          :class="barStyle"
          type="text"
          :placeholder="$t('search_bar.placeholder')"
          @keyup="(e) => {if (e.key === 'Enter') {userPressedSearchButtonOrEnter()} else {inputMightHaveChanged()}}"
          @focus="globalState.showDropDown = true"
        />
        <span
          class="searchbutton"
          :class="barStyle"
          @click="userPressedSearchButtonOrEnter()"
        >
          <FontAwesomeIcon :icon="faMagnifyingGlass" />
        </span>
      </div>
      <div v-if="globalState.showDropDown" ref="dropDown" class="drop-down" :class="barStyle">
        <div class="separation" :class="barStyle" />
        <div class="filter-area">
          <BcSearchbarNetworkSelector
            class="filter-networks"
            :initial-state="userFilters.networks"
            :bar-style="barStyle"
            @change="networkFilterHasChanged"
          />
          <BcSearchbarCategorySelectors
            class="filter-categories"
            :initial-state="userFilters.categories"
            :bar-style="barStyle"
            @change="categoryFilterHasChanged"
          />
        </div>
        <div v-if="globalState.state === States.ApiHasResponded" class="output-area" :class="barStyle">
          <div v-for="network of results.organized.in.networks" :key="network.chainId" class="network-container" :class="barStyle">
            <div v-for="typ of network.types" :key="typ.type" class="type-container" :class="barStyle">
              <BcSearchbarSuggestionLine
                v-for="(suggestion, i) of typ.suggestion"
                :key="i"
                :suggestion="suggestion"
                :chain-id="network.chainId"
                :result-type="typ.type"
                :bar-style="barStyle"
                @click="userClickedSuggestion"
              />
            </div>
          </div>
          <div v-if="results.organized.howManyResultsIn == 0" class="info center">
            {{ $t('search_bar.no_result_matches') }}
            {{ results.organized.howManyResultsOut > 0 ? $t('search_bar.your_filters') : $t('search_bar.your_input') }}
          </div>
          <div v-if="results.organized.howManyResultsOut > 0" class="info bottom">
            {{ (results.organized.howManyResultsIn == 0 ? ' (' : '+') + String(results.organized.howManyResultsOut) }}
            {{ (results.organized.howManyResultsOut == 1 ? $t('search_bar.result_hidden') : $t('search_bar.results_hidden')) +
              (results.organized.howManyResultsIn == 0 ? ')' : ' '+$t('search_bar.by_your_filters')) }}
          </div>
        </div>
        <div v-else class="output-area" :class="barStyle">
          <div v-if="globalState.state === States.InputIsEmpty" class="info center">
            {{ $t('search_bar.help') }}
          </div>
          <div v-else-if="globalState.state === States.SearchRequestWillBeSent || globalState.state === States.WaitingForResults" class="info center">
            {{ $t('search_bar.searching') }}
            <BcLoadingSpinner :loading="true" size="small" alignment="default" />
          </div>
          <div v-else-if="globalState.state === States.Error" class="info center">
            {{ $t('search_bar.something_wrong') }}
            <IconErrorFace :inline="true" />
            <br>
            {{ $t('search_bar.try_again') }}
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
  height: v-bind(searchButtonSize);
  &.discreet {
    @media (min-width: 600px) { // large screen
      width: 460px;
    }
    @media (max-width: 600px) { // mobile
      width: 380px;
    }
  }
  &.gaudy {
    @media (min-width: 600px) { // large screen
      width: 735px;
    }
    @media (max-width: 600px) { // mobile
      width: 380px;
    }
  }
}

.whole-component {
  @include main.container;
  position: absolute;
  left: 0px;
  right: 0px;
  &.dropdown-is-opened {
    z-index: 256;
  }

  &.discreet {
    background-color: var(--searchbar-background-discreet);
    border: none;
    &.dropdown-is-opened {
      border: 1px solid var(--searchbar-background-hover-discreet);
    }
  }
  &.gaudy {
    background-color: var(--searchbar-background-gaudy);
    border: 1px solid var(--input-border-color);
  }
}

.whole-component .input-and-button {
  display: flex;

  .input-field {
    display: flex;
    flex-grow: 1;
    left: 0;
    height: v-bind(searchButtonSize);
    border: none;
    box-shadow: none;
    background-color: transparent;
    color: var(--input-placeholder-text-color);
  }
  .searchbutton {
    display: flex;
    width: v-bind(searchButtonSize);
    height: v-bind(searchButtonSize);
    right: 0px;
    justify-content: center;
    align-items: center;
    border-radius: var(--border-radius);  // important because the button appears when hovered
    border: none;
    background-color: transparent;
    cursor: pointer;
    &.discreet {
      font-size: 15px;
      color: var(--input-placeholder-text-color);
      &:hover {
        background-color: var(--searchbar-background-hover-discreet);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
    &.gaudy {
      font-size: 18px;
      color: var(--text-color);
      &:hover {
        background-color: var(--dropdown-background-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }
}

.whole-component .drop-down {
  left: 0;
  right: 0;
  padding-left: 4px;
  padding-right: 4px;
  padding-bottom: 4px;
  .separation {
    left: 11px;
    right: 11px;
    height: 1px;
    margin-bottom: 10px;
    &.discreet {
      background-color: var(--searchbar-background-hover-discreet);
    }
    &.gaudy {
      background-color: var(--input-border-color);
    }
  }
}

.whole-component .drop-down .filter-area {
  display: flex;
  row-gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;

  .filter-networks {
    margin-left: 6px;
  }
}

.whole-component .drop-down .output-area {
  display: flex;
  flex-direction: column;
  min-height: 128px;
  max-height: 270px;  // the height of the filter section is subtracted
  right: 0px;
  overflow: auto;
  @include fonts.standard_text;
  &.discreet {
    color: var(--searchbar-text-discreet);
  }

  .network-container {
    display: flex;
    flex-direction: column;
    right: 0px;
    .type-container {
      display: flex;
      flex-direction: column;
      right: 0px;
    }
  }

  .info {
    width: 100%;
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
  }
}
</style>
