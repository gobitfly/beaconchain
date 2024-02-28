<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faMagnifyingGlass } from '@fortawesome/pro-solid-svg-icons'
import {
  Categories,
  CategoryInfo,
  ResultTypes,
  TypeInfo,
  getListOfResultTypes,
  convertSearchAheadResultIntoOrganizedResult,
  type SearchAheadResults,
  type OrganizedResults,
  type SearchBarStyle,
  type Matching
} from '~/types/searchengine'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const { t: $t } = useI18n()
const props = defineProps({
  searchable: { type: Array, required: true }, // list of categories that the bar can search in
  barStyle: { type: String, required: true }, // look of the bar ('discreet' for small, 'gaudy'  for big)
  pickByDefault: { type: Function, required: true } // when the user presses Enter, this callback function receives a simplified representation of the possible matches and must return one element from this list. The parameter (of type Matching[]) is a simplified view of the list of results sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. The bar will then trigger the event `@go` to call your handler with the result data of the matching that you picked.
})
const emit = defineEmits(['go'])

const barStyle : SearchBarStyle = props.barStyle as SearchBarStyle
const width : number = (barStyle === 'discreet' ? 460 : 735)
const height : number = (barStyle === 'discreet' ? 34 : 40)
const inputHeight = String(height) + 'px'
const inputWidth = String(width - height) + 'px'
const searchButtonSize = String(height) + 'px'

const searchable = props.searchable as Categories[]
let searchableTypes : ResultTypes[] = []

const PeriodOfDropDownUpdates = 500 /* TODO: change to 2000 for production !!!! */
const APIcallTimeout = 1500 // should not exceed PeriodOfDropDownUpdates

const waitingForSearchResults = ref(false)
const showDropDown = ref(false)
const populateDropDown = ref(true)
const inputted = ref('')
let lastKnownInput = ''
const networkDropdownOptions : {name: string, label: string}[] = []
const networkDropdownUserSelection = ref<string[]>([])
const inputFieldAndButton = ref<HTMLDivElement>()
const dropDown = ref<HTMLDivElement>()

const results = {
  raw: { data: [] } as SearchAheadResults, // response of the API, without structure nor order
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
  categories : Record<string, boolean>, // each field will have a Categories as key and the state of the button as value
  noCategoryIsSelected : boolean
}
const userFilters = ref<UserFilters>({
  networks: {},
  noNetworkIsSelected: true,
  everyNetworkIsSelected: false,
  categories: {},
  noCategoryIsSelected: true
})

function cleanUp () {
  lastKnownInput = ''
  inputted.value = ''
  waitingForSearchResults.value = false
  showDropDown.value = false
  populateDropDown.value = true
  results.raw = { data: [] }
}

onMounted(() => {
  searchableTypes = []
  // builds the list of search types from the list of searchable categories (obtained as a props)
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

  // listens to clicks outside the search engine
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
})

// In the V1, the server received a request between 1.5 and 3.5 seconds after the user inputted something, depending on the length of the input.
// Therefore, the average delay was ~2.5 s for the user as well as for the server. Most of the time the delay was shorter because the 3.5 s delay
// was only for entries of size 1.
// This less-than-2.5s-on-average delay arised from a Timeout Timer.
// For the V2, I propose to work with a 2-second Interval Timer because:
// - it makes sure that requests are not sent to the server more often than every 2 s (equivalent to V1),
// - while offering the user an average waiting time of 1 second through the magic of statistics (better than V1).
setInterval(() => {
  if (waitingForSearchResults.value) {
    if (searchAhead()) {
      filterAndOrganizeResults()
      waitingForSearchResults.value = false
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
    cleanUp()
  } else {
    waitingForSearchResults.value = true
    showDropDown.value = true
  }
}

function userFeelsLucky () {
  if (inputted.value.length === 0) {
    return
  }
  if (waitingForSearchResults.value) {
    if (!searchAhead()) {
      return
    }
  }
  filterAndOrganizeResults()
  if (results.organized.howManyResultsIn + results.organized.howManyResultsOut === 0) {
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
      // here we assume that the result with the best `closeness` value is the first one is array `type.found` (see the sorting done in `filterAnsdOrganizeResults()`)
      possibilities.push({ closeness: type.found[0].closeness, network: network.chainId, type: type.type })
    }
  }
  // calling back parent's function in charge of making a choice
  const picked = props.pickByDefault(possibilities)
  // retrieving the result corresponding to the choice
  const network = toConsider.networks.find(nw => nw.chainId === picked.network)
  const type = network?.types.find(ty => ty.type === picked.type)
  // calling back parent's function taking action with the result
  cleanUp()
  emit('go', type?.found[0].main, type?.type, network?.chainId)
}

function userClickedProposal (chain : ChainIDs, type : ResultTypes, found: string) {
  // cleans up and calls back user's function
  cleanUp()
  emit('go', found, type, chain)
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

function refreshDropDown () {
  populateDropDown.value = false
  filterAndOrganizeResults()
  populateDropDown.value = true // this triggers Vue to refresh the list of results
}

let searchAheadInProgress : boolean = false
// returns false if the API could not be reached or if it had a problem
// returns true otherwise (so also true when no result matches the input)
function searchAhead () : boolean {
  let error = false

  if (searchAheadInProgress) {
    return false
  }
  searchAheadInProgress = true

  // ********* SIMULATES AN API RESPONSE - TO BE REMOVED ONCE THE API IS IMPLEMENTED *********
  if (searchableTypes[0] as string !== '-- to be removed --') {
    results.raw = simulateAPIresponse(inputted.value)
  } else { // *** END OF STUFF TO REMOVE ***
    fetch('/api/2/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input: inputted.value, searchable: searchableTypes }),
      signal: AbortSignal.timeout(APIcallTimeout)
    }).then((received) => {
      if (received.ok && received.status < 400) {
        received.json().then((object) => {
          results.raw = object
        })
      } else {
        error = true
      }
    }).catch(() => {
      error = true
    })
    if (results.raw === undefined || results.raw.error !== undefined) {
      error = true
    }
  }

  if (error) {
    results.raw = { data: [] }
  }
  searchAheadInProgress = false
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
    const type = finding.type as ResultTypes

    // getting organized information from the finding
    const toBeAdded = convertSearchAheadResultIntoOrganizedResult(finding)
    if (toBeAdded.main === '') {
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
        found: []
      })
    }
    // now we can insert the finding at the right place in the organized results
    toBeAdded.closeness = calculateCloseness(toBeAdded.main)
    place.networks[existingNetwork].types[existingType].found.push(toBeAdded)
  }

  // This sorting orders the displayed results and is fundamental for function userFeelsLucky(). Do not alter the sorting without considering the needs of that function.
  function sortResults (place : OrganizedResults) {
    place.networks.sort((a, b) => ChainInfo[a.chainId].priority - ChainInfo[b.chainId].priority)
    for (const network of place.networks) {
      network.types.sort((a, b) => TypeInfo[a.type].priority - TypeInfo[b.type].priority)
      for (const type of network.types) {
        type.found.sort((a, b) => a.closeness - b.closeness)
      }
    }
  }
  sortResults(results.organized.in)
  sortResults(results.organized.out)
}

// Calculates how close the suggestion of result is to what the user typed.
// Guarantee: lower value <=> better matching
function calculateCloseness (suggestion : string) : number {
  // TODO ideally : calculate the Levenshtein distance between the two strings.
  // For now, the API suggests only exact matches and strings starting with the same letters as the user input. Therefore,
  // it is sufficient for now to return the length of the suggested result. This produces an ordering equivalent to
  // the ordering that the Levenshtein distance would produce.
  // Implementing the Levenshtein distance here would make SearchEngine.vue independent of any assumption about the back-end
  // search capabilities (in the future, the API might also give results approximating the input, for example with ENS names and graffiti).
  return suggestion.length - inputted.value.length
}

// ********* THIS FUNCTION SIMULATES AN API RESPONSE - TO BE REMOVED ONCE THE API IS IMPLEMENTED *********
function simulateAPIresponse (searched : string) : SearchAheadResults {
  const response : SearchAheadResults = {}; response.data = []

  // results are found 80% of the time
  if (Math.random() < 1 / 5.0) {
    return response
  }

  const n = Math.floor(Number(searched))
  const searchedIsPositiveInteger = (n !== Infinity && n >= 0 && String(n) === searched)

  response.data.push(
    {
      chain_id: 1,
      type: 'tokens',
      str_value: searched + 'Coin'
    },
    {
      chain_id: 1,
      type: 'addresses',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938'
    },
    {
      chain_id: 1,
      type: 'graffiti',
      str_value: searched + ' tutta la vita'
    }
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 1,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 1,
        type: 'slots',
        num_value: Number(searched)
      },
      {
        chain_id: 1,
        type: 'blocks',
        num_value: Number(searched)
      },
      {
        chain_id: 1,
        type: 'validators_by_index',
        num_value: Number(searched)
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 1,
        type: 'tokens',
        str_value: searched
      },
      {
        chain_id: 1,
        type: 'ens_names',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth'
      },
      {
        chain_id: 1,
        type: 'count_validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth',
        num_value: 7
      }
    )
  }
  response.data.push(
    {
      chain_id: 17000,
      type: 'addresses',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDEa7EA50F757Ca938'
    },
    {
      chain_id: 17000,
      type: 'count_validators_by_withdrawal_address',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A5357Ca938',
      num_value: 11
    },
    {
      chain_id: 42161,
      type: 'addresses',
      hash_value: '0x' + searched + '00000000000000000000000000CAFFE'
    },
    {
      chain_id: 42161,
      type: 'transactions',
      hash_value: '0x' + searched + 'a297ab886723ecfbc2cefab2ba385792058b344fbbc1f1e0a1139b2'
    },
    {
      chain_id: 8453,
      type: 'addresses',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa'
    },
    {
      chain_id: 8453,
      type: 'count_validators_by_deposit_address',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa',
      num_value: 150
    }
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 8453,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 8453,
        type: 'slots',
        num_value: Number(searched)
      },
      {
        chain_id: 8453,
        type: 'blocks',
        num_value: Number(searched)
      },
      {
        chain_id: 8453,
        type: 'validators_by_index',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'slots',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'blocks',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'validators_by_index',
        num_value: Number(searched)
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'USD'
      },
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + '42'
      },
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'Plus'
      },
      {
        chain_id: 100,
        type: 'tokens',
        str_value: searched + ' Coin'
      },
      {
        chain_id: 100,
        type: 'ens_names',
        str_value: searched + 'hallo.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 100,
        type: 'ens_names',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + 'hallo.eth'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth'
      },
      {
        chain_id: 100,
        type: 'count_validators_by_withdrawal_ens_name',
        str_value: searched + 'hallo.eth',
        num_value: 2
      },
      {
        chain_id: 100,
        type: 'count_validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth',
        num_value: 150
      }
    )
  }

  return response
}

// *** END OF THE FUNCTION TO BE REMOVED WHEN THE API IS IMPLEMENTED ***
</script>

<template>
  <div class="whole-engine">
    <div id="input-and-button" ref="inputFieldAndButton">
      <InputText
        id="input-field"
        v-model="inputted"
        :class="barStyle"
        type="text"
        placeholder="Search the blockchain"
        @keyup="(e) => {if (e.key === 'Enter') {userFeelsLucky()} else {inputMightHaveChanged()}}"
        @focus="showDropDown = inputted.length > 0"
      />
      <div
        id="searchbutton"
        :class="barStyle"
        @click="userFeelsLucky()"
      >
        <FontAwesomeIcon :icon="faMagnifyingGlass" />
      </div>
    </div>
    <div v-if="showDropDown" id="drop-down" ref="dropDown">
      <div id="filter-bar">
        <span id="filter-networks">
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
            @change="networkFilterHasChanged(); refreshDropDown()"
            @click="(e : Event) => e.stopPropagation()"
          />
          <div v-if="barStyle === 'discreet'" />
        </span>
        <label v-for="filter of Object.keys(userFilters.categories)" :key="filter" class="filter-button">
          <input
            v-model="userFilters.categories[filter]"
            class="hiddencheckbox"
            :true-value="true"
            :false-value="false"
            type="checkbox"
            @change="categoryFilterHasChanged(); refreshDropDown()"
          >
          <span class="face">{{ CategoryInfo[filter as Categories].filterLabel }}</span>
        </label>
      </div>
      <div v-if="waitingForSearchResults">
        {{ $t('search_engine.searching') }}
      </div>
      <div v-else-if="populateDropDown" id="panel-of-results">
        <div v-for="network of results.organized.in.networks" :key="network.chainId" class="network-container">
          <div class="network-title">
            <h2>{{ ChainInfo[network.chainId].name }}</h2>
          </div>
          <div v-for="types of network.types" :key="types.type" class="type-container">
            <div class="type-title">
              <h3>{{ TypeInfo[types.type].title }}</h3>
            </div>
            <div v-for="(found, i) of types.found" :key="i" class="single-result" @click="userClickedProposal(network.chainId, types.type, found.main)">
              {{ TypeInfo[types.type].preLabels }}
              {{ found.main }}
              <span v-if="found.complement !== ''">
                {{ TypeInfo[types.type].midLabels }}
                {{ found.complement }}
              </span>
              {{ TypeInfo[types.type].postLabels }}
            </div>
          </div>
        </div>
        <div id="absent-results">
          <span v-if="results.organized.howManyResultsIn == 0">
            {{ $t('search_engine.no_result_matches') }}
            {{ results.organized.howManyResultsOut > 0 ? $t('search_engine.your_filters') : $t('search_engine.your_input') }}
          </span>
          <span v-if="results.organized.howManyResultsOut > 0">
            {{ (results.organized.howManyResultsIn == 0 ? ' (' : '+') + String(results.organized.howManyResultsOut) }}
            {{ (results.organized.howManyResultsOut == 1 ? $t('search_engine.result_hidden') : $t('search_engine.results_hidden')) +
              (results.organized.howManyResultsIn == 0 ? ')' : ' '+$t('search_engine.by_your_filters')) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.whole-engine {
  position: relative;
}

#input-and-button {
  display: flex;

  #input-field {
    width: v-bind(inputWidth);
    height: v-bind(inputHeight);
    border-top-right-radius: 0px;
    border-bottom-right-radius: 0px;
    background-color: var(--searchbar-background);
    color: var(--text-color);
    box-shadow: none;
    &.discreet {
      border-color: var(--searchbar-background);
    }
    &.gaudy {
      border-color: var(--input-border-color);
    }
  }
  #searchbutton {
    display: flex;
    width: v-bind(searchButtonSize);
    height: v-bind(searchButtonSize);
    justify-content: center;
    align-items: center;
    border-top-left-radius: 0px;
    border-bottom-left-radius: 0px;
    border-top-right-radius: var(--border-radius);
    border-bottom-right-radius: var(--border-radius);
    cursor: pointer;
    &.discreet {
      background-color: var(--searchbar-background);
      font-size: 15px;
    }
    &.gaudy {
      background-color: var(--button-color-active);
      font-size: 18px;
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }
}

#drop-down {
  @include main.container;
  position: absolute;
  z-index: 256;
  left: 0;
  right: v-bind(searchButtonSize);
  background-color: var(--searchbar-background);
  padding: 4px;

  #panel-of-results {
    min-height: 200px;
    max-height: 300px;
    overflow: auto;
    h2 {
      margin: 0;
    }
    h3 {
      margin: 0;
    }
    .network-container {
      margin-bottom: 24px;
      .network-title {
        background-color: #b0b0b0;
        padding-left: 4px;
      }
      .type-container {
        border-bottom: 0.5px dashed var(--light-grey-3);
        padding: 4px;
        .type-title {

        }
        .single-result {
          cursor: pointer;
        }
      }
    }
  }
}

#drop-down #filter-bar {
  padding-top: 4px;
  padding-bottom: 8px;

  #filter-networks {
    margin-left: 6px;
    margin-right: 6px;
    margin-bottom: 6px;
    .p-multiselect {
      @include fonts.small_text_bold;
      width: 138px;
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
        }
      }
      &.p-multiselect-panel {
        width: 140px;
      }
    }
  }
  .filter-button {
    @include fonts.small_text_bold;
    .face{
      color: var(--primary-contrast-color);
      display: inline-block;
      border-radius: 10px;
      width: 75px;
      height: 17px;
      padding-top: 3px;
      text-align: center;
      margin-right: 6px;
      margin-bottom: 6px;
      transition: 0.2s;
      background-color: var(--button-color-disabled);
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

#absent-results {
  @include fonts.standard_text;
  color: var(--text-color-disabled);
  text-align: center;
}
</style>
