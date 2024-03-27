<script setup lang="ts">
import { warn } from 'vue'
import {
  Category,
  SubCategory,
  ResultType,
  FillFrom,
  type HowToFillresultSuggestionOutput,
  type ResultSuggestionOutput,
  CategoryInfo,
  SubCategoryInfo,
  TypeInfo,
  getListOfResultTypes,
  isOutputAnAPIresponse,
  type SearchAheadSingleResult,
  type SearchAheadResult,
  type ResultSuggestion,
  type OrganizedResults,
  type SearchBarStyle,
  SearchBarPurpose,
  type Matching,
  type PickingCallBackFunction
} from '~/types/searchbar'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const SearchRequestPeriodicity = 2 * 1000 // 2 seconds

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const props = defineProps<{
  searchable: Category[], // list of categories that the bar can search in
  unsearchable?: ResultType[], // the bar will not search for this types
  onlyNetworks?: ChainIDs[], // the bar will search on these networks only
  barStyle: SearchBarStyle, // look of the bar ('discreet', 'gaudy' or 'embedded')
  pickByDefault: PickingCallBackFunction /* When the user presses Enter, this callback function receives a simplified representation of
   the suggested results and returns one element from this list (or undefined). This list is passed in the parameter (of type Matching[])
   as a simplified view of the actual list of results. It is sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority.
   After you return a matching, the bar triggers the event `@go` to call your handler with the actual data of the result that you picked.
   If you return undefined instead of a matching, nothing happens (either no result suits you or you want to deactivate Enter). */
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

const barPurpose = whatIsMyPurpose()
let searchableTypes : ResultType[] = []
let allTypesBelongToAllNetworks = false

const inputted = ref('')
let lastKnownInput = ''
const globalState = ref<GlobalState>({
  state: States.InputIsEmpty,
  callAgainFunctionUserPressedSearchButtonOrEnter: false,
  showDropDown: false
})
const dropDown = ref<HTMLDivElement>()
const inputFieldAndButton = ref<HTMLDivElement>()

const zindexCorrectionClass = computed(() => globalState.value.showDropDown ? 'dropdown-is-opened' : '')

const userFilters = {
  networks: {} as Record<string, boolean>, // each field will have a stringifyEnum(ChainIDs) as key and the state of the option as value
  noNetworkIsSelected: true,
  categories: {} as Record<string, boolean>, // each field will have a stringifyEnum(Category) as key and the state of the button as value
  noCategoryIsSelected: true
}

const results = {
  raw: { data: [] } as SearchAheadResult, // response of the API, without structure nor order
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered-in results, organized
    howManyResultsIn: 0,
    out: { networks: [] } as OrganizedResults, // filtered-out results, organized
    howManyResultsOut: 0
  }
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
  allTypesBelongToAllNetworks = true
  // builds the list of all search types that the bar will consider, from the list of searchable categories (obtained as a props)
  for (const t of getListOfResultTypes(false)) {
    if (props.searchable.includes(TypeInfo[t].category) && !props.unsearchable?.includes(t)) {
      searchableTypes.push(t)
      allTypesBelongToAllNetworks &&= TypeInfo[t].belongsToAllNetworks // this variable will be used to know whether it is useless to show the network-filter selector
    }
  }
  // creates the fields storing the state of the category-filter buttons, and deselect them
  for (const s of props.searchable) {
    userFilters.categories[stringifyEnum(s)] = false
  }
  userFilters.noCategoryIsSelected = true
  // creates the fields storing the state of the network drop-down, and deselect all networks
  const networks = (props.onlyNetworks !== undefined && props.onlyNetworks.length > 0) ? props.onlyNetworks : getListOfImplementedChainIDs(true)
  for (const nw of networks) {
    userFilters.networks[stringifyEnum(nw)] = false
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
  if (!dropDown.value || !inputFieldAndButton.value ||
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
      body: {
        input: inputted.value,
        types: searchableTypes,
        count: isResultCountable(undefined)
      }
    })
  } catch (error) {
    received = undefined
  }
  if (inputted.value !== startInput) { // important: errors are ignored if outdated
    return ResultState.Outdated
  }
  if (!received || received.error !== undefined || received.data === undefined) {
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
  if (picked) {
    // retrieving the result corresponding to the choice
    const network = toConsider.networks.find(nw => nw.chainId === picked.network)
    const type = network?.types.find(ty => ty.type === picked.type)
    // calling back parent's function taking action with the result
    cleanUp(true)
    emit('go', type?.suggestion[0].queryParam, type?.type, network?.chainId, type?.suggestion[0].count)
  }
}

function userClickedSuggestion (chain : ChainIDs, type : ResultType, wanted: string, count : number) {
  // cleans up and calls back parent's function
  cleanUp(true)
  emit('go', wanted, type, chain, count)
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

function refreshOutputArea () {
  // updates the result lists with the latest API response and user filters
  filterAndOrganizeResults()
  // refreshes the output area in the drop-down
  updateGlobalState(globalState.value.state)
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
    const toBeAdded = convertOneSearchAheadResultIntoResultSuggestion(finding)
    if (!toBeAdded) {
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
    const stringifyedChainId = stringifyEnum(chainId)
    const stringifyedCategory = stringifyEnum(TypeInfo[type].category)
    const acceptTheChainID = (stringifyedChainId in userFilters.networks && (userFilters.networks[stringifyedChainId] || userFilters.noNetworkIsSelected)) || chainId === ChainIDs.Any
    const acceptTheCategory = stringifyedCategory in userFilters.categories && (userFilters.categories[stringifyedCategory] || userFilters.noCategoryIsSelected)
    if (acceptTheChainID && acceptTheCategory) {
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

// This function takes a single result element returned by the API and organizes it into an element simpler to handle by the
// code of the search bar because more... organized.
// If the result JSON from the API is somehow unexpected, the function returns `undefined`.
// The fields that the function reads in the API response as well as the place they are stored in our ResultSuggestion.output
// object are given by the filling information in TypeInfo[<result type>].howToFillresultSuggestionOutput in types/searchbar.ts
function convertOneSearchAheadResultIntoResultSuggestion (apiResponseElement : SearchAheadSingleResult) : ResultSuggestion | undefined {
  if (!(getListOfResultTypes(false) as string[]).includes(apiResponseElement.type)) {
    warn('The API returned an unexpected type of search-ahead result: ', apiResponseElement.type)
    return undefined
  }

  const type = apiResponseElement.type as ResultType
  const howToFillresultSuggestionOutput = TypeInfo[type].howToFillresultSuggestionOutput
  const output = {} as ResultSuggestionOutput

  for (const k in howToFillresultSuggestionOutput) {
    const key = k as keyof HowToFillresultSuggestionOutput
    const data = getAPIdata(apiResponseElement, howToFillresultSuggestionOutput[key])
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
    output.name = TypeInfo[type].title
    nameWasUnknown = true
  }

  // retrieving the data that identifies this very result in the back-end (will be given to the callback function `@go`)
  const queryParam = getAPIdata(apiResponseElement, TypeInfo[type].queryParamField) as string

  // Getting the number of identical results found. If the API did not clarify the number results for a countable type, we give NaN.
  let count = 1
  if (isResultCountable(type)) {
    count = (apiResponseElement.num_value === undefined) ? NaN : apiResponseElement.num_value
  }

  // We calculate how far the user input is from the result suggestion of the API (the API completes/approximates inputs, for example for graffiti).
  // It will be needed later to pick the best result suggestion when the user hits Enter, and also in the drop-down to order the suggestions by relevance when several results exist in a type group
  let closeness = Number.MAX_SAFE_INTEGER
  for (const k in output) {
    const key = k as keyof HowToFillresultSuggestionOutput
    if (isOutputAnAPIresponse(type, key)) {
      const cl = resemblanceWithInput(output[key])
      if (cl < closeness) {
        closeness = cl
      }
    }
  }

  return { output, nameWasUnknown, queryParam, closeness, count }
}

function getAPIdata (apiResponseElement : SearchAheadSingleResult, fillFrom : FillFrom | string) : string | undefined {
  const type = apiResponseElement.type as ResultType
  let sourceField : keyof SearchAheadSingleResult

  switch (fillFrom) {
    case FillFrom.SASRstr_value :
      sourceField = 'str_value'
      break
    case FillFrom.SASRnum_value :
      sourceField = 'num_value'
      break
    case FillFrom.SASRhash_value :
      sourceField = 'hash_value'
      break
    case FillFrom.CategoryTitle :
      return CategoryInfo[TypeInfo[type].category].title
    case FillFrom.SubCategoryTitle :
      return SubCategoryInfo[TypeInfo[type].subCategory].title
    case FillFrom.TypeTitle :
      return TypeInfo[type].title
    default :
      return fillFrom as string // this is a predefined text, hard-coded in object TypeInfo defined in searchbar.ts
  }

  if (apiResponseElement[sourceField] !== undefined) {
    return String(apiResponseElement[sourceField])
  }

  return undefined
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

function isResultCountable (type : ResultType | undefined) : boolean {
  if (type !== undefined) {
    return TypeInfo[type].countable
  }
  // from here, there is uncertainty but we must simply tell whether counting is possible for some results
  if (barPurpose === SearchBarPurpose.General) {
    return false // we do not ask the API to count identical results when the bar is versatile (general bar to search anything on the blockchain)
  }
  for (const type of searchableTypes) {
    if (TypeInfo[type].countable) {
      return true
    }
  }
  return false
}

function whatIsMyPurpose () : SearchBarPurpose {
  if (props.barStyle === 'embedded') {
    if (props.searchable.length === 1) {
      if (props.searchable[0] === Category.Addresses) {
        return SearchBarPurpose.Accounts
      } else if (props.searchable[0] === Category.Validators) {
        return SearchBarPurpose.Validators
      }
    }
    // there is no reason to reach this state but let's be careful:
    warn('The purpose of the search bar could not be determined')
  }

  return SearchBarPurpose.General
}

function mustNetworkFilterBeShown () : boolean {
  return Object.keys(userFilters.networks).length >= 2 && !allTypesBelongToAllNetworks
}

function mustCategoryFiltersBeShown () : boolean {
  return Object.keys(userFilters.categories).length >= 2
}

function inputPlaceHolder () : string {
  switch (barPurpose) {
    case SearchBarPurpose.General : return $t('search_bar.general_placeholder')
    case SearchBarPurpose.Accounts : return $t('search_bar.account_placeholder')
    case SearchBarPurpose.Validators : return $t('search_bar.validator_placeholder')
  }
  return '' // cannot happen but the static analysis thinks it can
}

function informationIfInputIsEmpty () : string {
  const info = $t('search_bar.type_something') + ' '

  switch (barPurpose) {
    case SearchBarPurpose.General : return info + $t('search_bar.and_use_filters')
    case SearchBarPurpose.Accounts : return info + $t('search_bar.related_to_account')
    case SearchBarPurpose.Validators : return info + $t('search_bar.related_to_validator')
  }
  return info // cannot happen but the static analysis thinks it can
}

function areThereResultsHiddenByUser () : boolean {
  return (mustNetworkFilterBeShown() || mustCategoryFiltersBeShown()) && results.organized.howManyResultsOut > 0
}

function informationIfNoResult () : string {
  let info = $t('search_bar.no_result_matches') + ' '

  if (areThereResultsHiddenByUser()) {
    info += $t('search_bar.your_filters')
  } else {
    info += $t('search_bar.your_input')
  }

  return info
}

function informationIfHiddenResults () : string {
  let info = String(results.organized.howManyResultsOut) + ' '

  info += (results.organized.howManyResultsOut === 1 ? $t('search_bar.one_result_hidden') : $t('search_bar.several_results_hidden'))

  if (results.organized.howManyResultsIn !== 0) {
    info = '+' + info + ' ' + $t('search_bar.by_your_filters')
  } else {
    info = '(' + info + ')'
  }

  return info
}

// This padding with leading zeros is required otherwise the fields in a Record whose keys are
// enum values would get sorted lexicographically for some reason.
function stringifyEnum (enumValue : Category | SubCategory | ChainIDs) : string {
  return String(enumValue).padStart(10, '0')
}
</script>

<template>
  <div class="anchor" :class="barStyle">
    <div class="whole-component" :class="[barStyle, zindexCorrectionClass]">
      <div ref="inputFieldAndButton" class="input-and-button" :class="barStyle">
        <InputText
          v-model="inputted"
          class="input-field"
          :class="barStyle"
          type="text"
          :placeholder="inputPlaceHolder()"
          @keyup="(e) => {if (e.key === 'Enter') {userPressedSearchButtonOrEnter()} else {inputMightHaveChanged()}}"
          @focus="globalState.showDropDown = true"
        />
        <BcSearchbarButton
          class="searchbutton"
          :class="[barStyle, zindexCorrectionClass]"
          :bar-style="barStyle"
          :bar-purpose="barPurpose"
          @click="userPressedSearchButtonOrEnter()"
        />
      </div>
      <div v-if="globalState.showDropDown" ref="dropDown" class="drop-down" :class="barStyle">
        <div class="separation" :class="barStyle" />
        <div v-if="mustNetworkFilterBeShown() || mustCategoryFiltersBeShown()" class="filter-area">
          <BcSearchbarNetworkSelector
            v-if="mustNetworkFilterBeShown()"
            class="filter-networks"
            :initial-state="userFilters.networks"
            :bar-style="barStyle"
            @change="networkFilterHasChanged"
          />
          <BcSearchbarCategorySelectors
            v-if="mustCategoryFiltersBeShown()"
            class="filter-categories"
            :initial-state="userFilters.categories"
            :bar-style="barStyle"
            @change="categoryFilterHasChanged"
          />
        </div>
        <div v-if="globalState.state === States.ApiHasResponded" class="output-area" :class="barStyle">
          <div v-for="network of results.organized.in.networks" :key="network.chainId" class="network-container" :class="barStyle">
            <div v-for="typ of network.types" :key="typ.type" class="type-container" :class="barStyle">
              <div v-for="(suggestion, i) of typ.suggestion" :key="i" class="suggestionrow-container" :class="barStyle">
                <BcSearchbarSuggestionRow
                  :suggestion="suggestion"
                  :chain-id="network.chainId"
                  :result-type="typ.type"
                  :bar-style="barStyle"
                  :bar-purpose="barPurpose"
                  @row-selected="userClickedSuggestion"
                />
                <div class="separation-between-suggestions" :class="barStyle" />
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
          <div v-if="globalState.state === States.InputIsEmpty" class="info center">
            {{ informationIfInputIsEmpty() }}
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
  &.embedded {
    height: 30px;
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
  left: 0px;
  right: 0px;

  &.gaudy,
  &.embedded {
    background-color: var(--searchbar-background-gaudy);
    border: 1px solid var(--input-border-color);
  }
  &.discreet {
    background-color: var(--searchbar-background-discreet);
    border: none;
    &.dropdown-is-opened {
      border: 1px solid var(--searchbar-background-hover-discreet);
    }
  }
}

.whole-component .input-and-button {
  display: block;
  width: 100%;

  .input-field {
    left: 0;
    width: 100%;
    border: none;
    box-shadow: none;
    background-color: transparent;
    color: var(--input-placeholder-text-color);
    &.gaudy {
      height: 40px;
      padding-right: 41px;
    }
    &.discreet {
      height: 34px;
      padding-right: 35px;
    }
    &.embedded {
      height: 30px;
      padding-right: 31px;
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

.whole-component .drop-down {
  position: relative;
  width: 100%;
  padding-bottom: 4px;

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
    row-gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 8px;

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
    min-height: 128px;
    max-height: 270px;  // the height of the filter section is subtracted
    overflow: auto;
    @include fonts.standard_text;
    &.discreet {
      color: var(--searchbar-text-discreet);
    }

    .network-container {
      position: relative;
      display: flex;
      flex-direction: column;
      width: 100%;

      .type-container {
        position: relative;
        display: flex;
        flex-direction: column;
        width: 100%;

        .suggestionrow-container {
          position: relative;
          width: 100%;

          .separation-between-suggestions {
            position: relative;
            display: none;
            margin-left: 8px;
            margin-right: 8px;
            height: 0.9px;

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
}
</style>
