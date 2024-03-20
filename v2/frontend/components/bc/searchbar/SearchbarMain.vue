<script setup lang="ts">
import { warn } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faMagnifyingGlass } from '@fortawesome/pro-solid-svg-icons'
import {
  Category,
  CategoryInfo,
  ResultType,
  TypeInfo,
  getListOfResultTypes,
  getListOfResultTypesInCategory,
  type SearchAheadSingleResult,
  type SearchAheadResult,
  type SearchBarStyle,
  type Matching
} from '~/types/searchbar'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const { t: $t } = useI18n()
const props = defineProps({
  searchable: { type: Array, required: true }, // list of categories that the bar can search in
  barStyle: { type: String, required: true }, // look of the bar ('discreet', 'gaudy' or 'embedded')
  pickByDefault: { type: Function, required: true } // when the user presses Enter, this callback function receives a simplified representation of the possible matches and must return one element from this list. The parameter (of type Matching[]) is a simplified view of the list of results sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. The bar will then trigger the event `@go` to call your handler with the result data of the matching that you picked.
})
const emit = defineEmits(['go'])

enum States {
  InputIsEmpty,
  WaitingForResults,
  ApiHasResponded,
  Error,
  UpdateIncoming
}
interface SearchState {
  state : States,
  numberOfApiCallsWithoutResponse : number,
  userFeelsLucky: boolean
}

interface ResultSuggestion {
  columns: string[],
  queryParam: number, // index of the string given to the callback function `@go`
  closeness: number // how close the suggested result is to the user input (important for graffiti, later for other things if the back-end evolves to find other approximate results)
}
interface OrganizedResults {
  networks: {
    chainId: ChainIDs,
    types: {
      type: ResultType,
      suggestion: ResultSuggestion[]
    }[]
  }[]
}

const PeriodOfDropDownUpdates = 2000
const NumberOfApiCallAttemptsBeforeShowingError = 2

const barStyle : SearchBarStyle = props.barStyle as SearchBarStyle
const searchButtonSize = (barStyle === 'discreet') ? '34px' : '40px'

const searchable = props.searchable as Category[]
let searchableTypes : ResultType[] = []

const inputted = ref('')
let lastKnownInput = ''
const searchState = ref<SearchState>({
  state: States.InputIsEmpty,
  numberOfApiCallsWithoutResponse: 0,
  userFeelsLucky: false
})
const showDropDown = ref<boolean>(false)
const networkDropdownOptions : {name: string, label: string}[] = []
const networkDropdownUserSelection = ref<string[]>([])
const dropDown = ref<HTMLDivElement>()
const inputFieldAndButton = ref<HTMLDivElement>()

const results = {
  raw: { data: [] } as SearchAheadResult, // response of the API, without structure nor order
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered results, organized
    howManyResultsIn: 0,
    out: { networks: [] } as OrganizedResults, // filtered out results, organized
    howManyResultsOut: 0
  }
}

interface UserFilters {
  networks: Record<string, boolean>, // each field will have a String(ChainIDs) as key and the state of the option as value
  noNetworkIsSelected : boolean,
  everyNetworkIsSelected : boolean,
  categories : Record<string, boolean>, // each field will have a Category as key and the state of the button as value
  noCategoryIsSelected : boolean
}
const userFilters = ref<UserFilters>({
  networks: {},
  noNetworkIsSelected: true,
  everyNetworkIsSelected: false,
  categories: {},
  noCategoryIsSelected: true
})

function cleanUp (closeDropDown : boolean) {
  lastKnownInput = ''
  inputted.value = ''
  resetSearchState()
  if (closeDropDown) {
    showDropDown.value = false // not equivalent to `showDropDown.value = !closeDropDown` because it must not be opened when it is already closed
  }
  results.raw = { data: [] }
}

function resetSearchState (state : States = States.InputIsEmpty) {
  if (state === searchState.value.state) {
    // makes sure that Vue re-renders the drop-down although the state does not change
    searchState.value.state = States.UpdateIncoming
  }
  searchState.value.numberOfApiCallsWithoutResponse = 0
  searchState.value.userFeelsLucky = false
  searchState.value.state = state
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
    userFilters.value.categories[s] = false
  }
  userFilters.value.noCategoryIsSelected = true

  for (const nw of getListOfImplementedChainIDs(true)) {
    // creates the field telling us whether this network is selected
    userFilters.value.networks[String(nw)] = false
    // populates the network-drop-down
    networkDropdownOptions.push({ name: String(nw), label: ChainInfo[nw].name })
  }
  networkDropdownUserSelection.value = [] // deselects all options
  networkFilterHasChanged()

  // listens to clicks outside the component
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
})

// In the V1, the server received a request between 1.5 and 3.5 seconds after the user inputted something, depending on the length of the input.
// Therefore, the average delay was ~2.5 s for the user as well as for the server. Most of the time the delay was shorter because the 3.5 s delay
// was only for entries of size 1.
// This less-than-2.5s-on-average delay arised from a Timeout Timer.
// For the V2, I propose to work with an Interval Timer because:
// - it makes sure that requests are not sent to the server more often than every 2 s (equivalent to V1),
// - while offering the user an average waiting time of only 1 second through the magic of statistics (better than V1).
setInterval(() => {
  if (searchState.value.state === States.WaitingForResults) {
    if (!searchAhead()) {
      // the communication with the API failed or the API is down
      searchState.value.numberOfApiCallsWithoutResponse++
      if (searchState.value.numberOfApiCallsWithoutResponse >= NumberOfApiCallAttemptsBeforeShowingError) {
        resetSearchState(States.Error)
      }
    } else {
      const callFunctionUserFeelsLucky = searchState.value.userFeelsLucky // this value must be retrieved now because of the call to resetSearchState() before it is used
      filterAndOrganizeResults()
      resetSearchState(States.ApiHasResponded)
      if (callFunctionUserFeelsLucky) {
        userFeelsLucky()
      }
    }
  }
},
PeriodOfDropDownUpdates
)

// closes the drop-down if the user interacts with another part of the page
function listenToClicks (event : Event) {
  if (dropDown.value === undefined || inputFieldAndButton.value === undefined ||
      dropDown.value.contains(event.target as Node) || inputFieldAndButton.value.contains(event.target as Node)) {
    return
  }
  showDropDown.value = false
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
    resetSearchState(States.WaitingForResults)
  }
}

function userFeelsLucky () {
  if (searchState.value.state === States.InputIsEmpty) {
    return
  }
  // if the previous API call failed and the user tries again with Enter or the search button
  if (searchState.value.state === States.Error) {
    // we order a new search (the timer will lanuch it)
    resetSearchState(States.WaitingForResults)
  }
  // if we are waiting for a response from the API (because of inputMightHaveChanged() or because of the retry just above)
  if (searchState.value.state === States.WaitingForResults) {
    // we ask the timer to call this function once (and if) the results are received
    searchState.value.userFeelsLucky = true
    // in the meantime, we do not proceed further
    return
  }

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

function userClickedProposal (chain : ChainIDs, type : ResultType, what: string) {
  // cleans up and calls back user's function
  cleanUp(true)
  emit('go', what, type, chain)
}

function networkFilterHasChanged () {
  userFilters.value.noNetworkIsSelected = (networkDropdownUserSelection.value.length === 0)
  userFilters.value.everyNetworkIsSelected = (networkDropdownUserSelection.value.length === networkDropdownOptions.length)

  for (const nw in userFilters.value.networks) {
    userFilters.value.networks[nw] = networkDropdownUserSelection.value.includes(nw)
  }
}

function categoryFilterHasChanged () {
  // determining whether any filter button is activated
  let allButtonsOff = true
  for (const cat in userFilters.value.categories) {
    if (userFilters.value.categories[cat]) {
      allButtonsOff = false
      break
    }
  }
  userFilters.value.noCategoryIsSelected = allButtonsOff
}

// returns false if the API could not be reached or if it had a problem
// returns true otherwise (so also true when no result matches the input)
function searchAhead () : boolean {
  let error = false

  useCustomFetch<SearchAheadResult>(API_PATH.SEARCH, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: { input: inputted.value, searchable: searchableTypes }
  }).then((received) => {
    results.raw = received
  }).catch(() => {
    error = true
  })
  if (results.raw === undefined || results.raw.error !== undefined) {
    error = true
  }

  if (error) {
    results.raw = { data: [] }
  }
  return !error
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
    // note that when the user did not select any network or any category, we default to showing all of them
    let place : OrganizedResults
    if ((userFilters.value.networks[String(chainId)] || userFilters.value.noNetworkIsSelected || chainId === ChainIDs.Any) &&
        (userFilters.value.categories[TypeInfo[type].category] || userFilters.value.noCategoryIsSelected)) {
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

  // This sorting orders the displayed results and is fundamental for function userFeelsLucky(). Do not alter the sorting without considering the needs of that function.
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

function filterHint (category : Category) : string {
  let hint = $t('search_bar.shows') + ' '

  if (category === Category.Validators) {
    hint += $t('search_bar.this_type') + ' '
    hint += 'Validator'
  } else {
    const list = getListOfResultTypesInCategory(category, false)

    hint += (list.length === 1 ? $t('search_bar.this_type') : $t('search_bar.these_types')) + ' '
    for (let i = 0; i < list.length; i++) {
      hint += TypeInfo[list[i]].title
      if (i < list.length - 1) {
        hint += ', '
      }
    }
  }

  return hint
}

function refreshOutputArea () {
  // updates the result lists with the latest API response and user filters
  filterAndOrganizeResults()
  // refreshes the output area in the drop-down
  resetSearchState(searchState.value.state)
}
</script>

<template>
  <div id="anchor" :class="barStyle">
    <div id="whole-component" :class="[barStyle, showDropDown?'dropdown-is-opened':'']">
      <div id="input-and-button" ref="inputFieldAndButton" :class="barStyle">
        <InputText
          id="input-field"
          v-model="inputted"
          :class="barStyle"
          type="text"
          :placeholder="$t('search_bar.placeholder')"
          @keyup="(e) => {if (e.key === 'Enter') {userFeelsLucky()} else {inputMightHaveChanged()}}"
          @focus="showDropDown = true"
        />
        <span
          id="searchbutton"
          :class="barStyle"
          @click="userFeelsLucky()"
        >
          <FontAwesomeIcon :icon="faMagnifyingGlass" />
        </span>
      </div>
      <div v-if="showDropDown" id="drop-down" ref="dropDown" :class="barStyle">
        <div id="separation" :class="barStyle" />
        <div id="filter-area">
          <div id="filter-networks">
            <!--do not remove '&nbsp;' in the placeholder otherwise the CSS of the component believes that nothing is selected when everthing is selected-->
            <MultiSelect
              v-model="networkDropdownUserSelection"
              :options="networkDropdownOptions"
              option-value="name"
              option-label="label"
              placeholder="Networks:&nbsp;all"
              :variant="'filled'"
              display="comma"
              :show-toggle-all="false"
              :max-selected-labels="1"
              :selected-items-label="'Networks: ' + (userFilters.everyNetworkIsSelected ? 'all' : '{0}')"
              append-to="self"
              @change="networkFilterHasChanged(); refreshOutputArea()"
              @click="(e : Event) => e.stopPropagation()"
            />
          </div>
          <div id="filter-categories">
            <span v-for="filter of Object.keys(userFilters.categories)" :key="filter">
              <BcTooltip :text="filterHint(filter as Category)">
                <label class="filter-button">
                  <input
                    v-model="userFilters.categories[filter]"
                    class="hiddencheckbox"
                    :true-value="true"
                    :false-value="false"
                    type="checkbox"
                    @change="categoryFilterHasChanged(); refreshOutputArea()"
                  >
                  <span class="face" :class="barStyle">
                    {{ CategoryInfo[filter as Category].filterLabel }}
                  </span>
                </label>
              </BcTooltip>
            </span>
          </div>
        </div>
        <div v-if="searchState.state === States.ApiHasResponded" class="output-area" :class="barStyle">
          <div v-for="network of results.organized.in.networks" :key="network.chainId" class="network-container" :class="barStyle">
            <div v-for="typ of network.types" :key="typ.type" class="type-container" :class="barStyle">
              <div
                v-for="(suggestion, i) of typ.suggestion"
                :key="i"
                class="single-result"
                :class="barStyle"
                @click="userClickedProposal(network.chainId, typ.type, suggestion.columns[suggestion.queryParam])"
              >
                <span v-if="network.chainId !== ChainIDs.Any" class="columns-icons" :class="barStyle">
                  <BcSearchbarTypeIcons :type="typ.type" class="type-icon not-alone" />
                  <IconNetwork :chain-id="network.chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
                </span>
                <span v-else class="columns-icons" :class="barStyle">
                  <BcSearchbarTypeIcons :type="typ.type" class="type-icon alone" />
                </span>
                <span class="columns-0" :class="barStyle">
                  <BcSearchbarMiddleEllipsis>{{ suggestion.columns[0] }}</BcSearchbarMiddleEllipsis>
                </span>
                <span class="columns-1and2" :class="barStyle">
                  <span v-if="suggestion.columns[1] !== ''" class="columns-1" :class="barStyle">
                    <BcSearchbarMiddleEllipsis>{{ suggestion.columns[1] }}</BcSearchbarMiddleEllipsis>
                  </span>
                  <span v-if="suggestion.columns[2] !== ''" class="columns-2" :class="[barStyle,(suggestion.columns[1] !== '')?'greyish':'']">
                    <BcSearchbarMiddleEllipsis v-if="TypeInfo[typ.type].dropdownColumns[1] === undefined" :width-is-fixed="true">({{ suggestion.columns[2] }})</BcSearchbarMiddleEllipsis>
                    <BcSearchbarMiddleEllipsis v-else :width-is-fixed="true">{{ suggestion.columns[2] }}</BcSearchbarMiddleEllipsis>
                  </span>
                </span>
                <span class="columns-category" :class="barStyle">
                  <span class="category-label" :class="barStyle">
                    {{ CategoryInfo[TypeInfo[typ.type].category].filterLabel }}
                  </span>
                </span>
              </div>
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
          <div v-if="searchState.state === States.InputIsEmpty" class="info center">
            {{ $t('search_bar.help') }}
          </div>
          <div v-else-if="searchState.state === States.WaitingForResults" class="info center">
            {{ $t('search_bar.searching') }}
            <BcLoadingSpinner :loading="true" size="small" alignment="default" />
          </div>
          <div v-else-if="searchState.state === States.Error" class="info center">
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

#anchor {
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

#whole-component {
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

#whole-component #input-and-button {
  display: flex;

  #input-field {
    display: flex;
    flex-grow: 1;
    left: 0;
    height: v-bind(searchButtonSize);
    border: none;
    box-shadow: none;
    background-color: transparent;
    color: var(--input-placeholder-text-color);
  }
  #searchbutton {
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

#whole-component #drop-down {
  left: 0;
  right: 0;
  padding-left: 4px;
  padding-right: 4px;
  padding-bottom: 4px;
  #separation {
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

#whole-component #drop-down #filter-area {
  display: flex;
  row-gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;

  #filter-networks {
    margin-left: 6px;

    .p-multiselect {
      @include fonts.small_text_bold;
      width: 128px;
      height: 20px;
      border-radius: 10px;
      .p-multiselect-trigger {
        width: 1.5rem;
      }
      .p-multiselect-label {
        padding-top: 3px;
        border-top-left-radius: 10px;
        border-bottom-left-radius: 10px;
        .p-placeholder {
          border-top-left-radius: 10px;
          border-bottom-left-radius: 10px;
          background: var(--searchbar-filter-unselected-gaudy);
        }
      }
      &.p-multiselect-panel {
        width: 140px;
        max-height: 100px;
        overflow: auto;
      }
    }
  }
  .filter-button {
    @include fonts.small_text_bold;

    .face{
      display: inline-block;
      border-radius: 10px;
      height: 17px;
      padding-top: 2.5px;
      padding-left: 8px;
      padding-right: 8px;
      text-align: center;
      margin-left: 6px;
      transition: 0.2s;
      &.discreet {
        color: var(--light-black);
        background-color: var(--light-grey);
      }
      &.gaudy {
        color: var(--primary-contrast-color);
        background-color: var(--searchbar-filter-unselected-gaudy);
      }
      &:hover {
        background-color: var(--light-grey-3);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
    .hiddencheckbox {
      display: none;
      width: 0;
      height: 0;
    }
    .hiddencheckbox:checked + .face {
      background-color: var(--button-color-active);
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }
}

#whole-component #drop-down .output-area {
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
      .single-result {
        cursor: pointer;
        display: grid;
        min-width: 0;
        right: 0px;
        padding-top: 7px;
        padding-bottom: 7px;
        @media (min-width: 600px) { // large screen
          &.gaudy {
            grid-template-columns: 40px 106px 488px auto;
            padding-left: 4px;
            padding-right: 4px;
          }
          &.discreet {
            grid-template-columns: 40px 106px 298px;
          }
        }
        @media (max-width: 600px) { // mobile
          grid-template-columns: 40px 106px 218px;
        }
        border-radius: var(--border-radius);

        &:hover {
          &.discreet {
            background-color: var(--searchbar-background-hover-discreet);
          }
          &.gaudy {
            background-color: var(--dropdown-background-hover);
          }
        }
        &:active {
          &.discreet {
            background-color: var(--searchbar-background-pressed-discreet);
          }
          &.gaudy {
            background-color: var(--button-color-pressed);
          }
        }

        .columns-icons {
          position: relative;
          grid-column: 1;
          grid-row: 1;
          @media (max-width: 600px) { // mobile
            grid-row-end: span 2;
          }
          &.discreet {
            grid-row-end: span 2;
          }
          display: flex;
          margin-top: auto;
          margin-bottom: auto;
          width: 30px;
          height: 36px;

          .type-icon {
            &.not-alone {
              display: inline;
              position: relative;
              top: 2px;
            }
            &.alone {
             display: flex;
             margin-top: auto;
             margin-bottom: auto;
            }
            width: 20px;
            max-height: 20px;
          }
          .network-icon {
            position: absolute;
            bottom: 0px;
            right: 0px;
            width: 20px;
            height: 20px;
          }
        }
        .columns-0 {
          grid-column: 2;
          grid-row: 1;
          display: inline-block;
          position: relative;
          margin-top: auto;
          &.gaudy {
            margin-bottom: auto;
          }
          margin-right: 14px;
          left: 0px;
          font-weight: var(--roboto-medium);
        }
        .columns-1and2 {
          grid-column: 3;
          grid-row: 1;
          display: flex;
          @media (max-width: 600px) { // mobile
            grid-row-end: span 2;
            flex-direction: column;
          }
          &.discreet {
            grid-row-end: span 2;
            flex-direction: column;
          }
          position: relative;
          margin-top: auto;
          margin-bottom: auto;
          left: 0px;
          font-weight: var(--roboto-medium);
          white-space: nowrap;

          .columns-1 {
            display: flex;
            max-width: 100%;
            @media (min-width: 600px) { // large screen
              &.gaudy {
                max-width: 27%;
              }
            }
            position: relative;
            margin-right: 0.8ch;
          }
          .columns-2 {
            display: flex;
            position: relative;
            flex-grow: 1;
            &.greyish.discreet {
              color: var(--searchbar-text-detail-discreet);
            }
            &.greyish.gaudy {
              color: var(--searchbar-text-detail-gaudy);
            }
          }
        }
        .columns-category {
          display: block;
          @media (min-width: 600px) { // large screen
            &.gaudy {
              grid-column: 4;
              grid-row: 1;
              margin-top: auto;
              margin-bottom: auto;
              margin-right: 2px;
              float: right;
              justify-content: right;
              text-align: right;
            }
            &.discreet {
              grid-column: 2;
              grid-row: 2;
            }
          }
          @media (max-width: 600px) { // mobile
            grid-column: 2;
            grid-row: 2;
          }
          .category-label {
            &.discreet {
              color: var(--searchbar-text-detail-discreet);
            }
            &.gaudy {
              color: var(--searchbar-text-detail-gaudy);
            }
          }
        }
      }
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
