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
} from '~/types/searchengine'
import { ChainIDs, ChainInfo, getListOfImplementedChainIDs } from '~/types/networks'

const { t: $t } = useI18n()
const props = defineProps({
  searchable: { type: Array, required: true }, // list of categories that the bar can search in
  barStyle: { type: String, required: true }, // look of the bar ('discreet' for small, 'gaudy'  for big)
  pickByDefault: { type: Function, required: true } // when the user presses Enter, this callback function receives a simplified representation of the possible matches and must return one element from this list. The parameter (of type Matching[]) is a simplified view of the list of results sorted by ChainInfo[chainId].priority and TypeInfo[resultType].priority. The bar will then trigger the event `@go` to call your handler with the result data of the matching that you picked.
})
const emit = defineEmits(['go'])

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

const barStyle : SearchBarStyle = props.barStyle as SearchBarStyle
const searchButtonSize = (barStyle === 'discreet') ? '34px' : '40px'

const searchable = props.searchable as Category[]
let searchableTypes : ResultType[] = []

const PeriodOfDropDownUpdates = 2000
const APIcallTimeout = 1500 // should not exceed PeriodOfDropDownUpdates

const waitingForSearchResults = ref(false)
const numberOfApiCallsWithoutResponse = ref(0)
const showDropDown = ref(false)
const populateDropDown = ref(true)
const inputted = ref('')
let lastKnownInput = ''
const networkDropdownOptions : {name: string, label: string}[] = []
const networkDropdownUserSelection = ref<string[]>([])
const inputFieldAndButton = ref<HTMLDivElement>()
const dropDown = ref<HTMLDivElement>()

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

function cleanUp () {
  lastKnownInput = ''
  inputted.value = ''
  waitingForSearchResults.value = false
  numberOfApiCallsWithoutResponse.value = 0
  populateDropDown.value = false
  results.raw = { data: [] }
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
    if (!searchAhead()) {
      numberOfApiCallsWithoutResponse.value++
      // `waitingForSearchResults.value` remains true so we will try again in 2 seconds
    } else {
      filterAndOrganizeResults()
      waitingForSearchResults.value = false
      numberOfApiCallsWithoutResponse.value = 0
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
    populateDropDown.value = true
  }
}

function userFeelsLucky () {
  if (inputted.value.length === 0) {
    return
  }
  if (waitingForSearchResults.value) {
    // the timer did not trigger a search yet, so we do it
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
  cleanUp()
  emit('go', type?.suggestion[0].columns[type?.suggestion[0].queryParam], type?.type, network?.chainId)
}

function userClickedProposal (chain : ChainIDs, type : ResultType, what: string) {
  // cleans up and calls back user's function
  cleanUp()
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
    if (Math.random() < 1 / 2.5) {
      // 40% of the time, we simulate an error (the timer will try again)
      error = true
    }
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
// in the drop-down are set in the object `TypeInfo` filled in types/searchengine.ts, by its properties
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
  let hint : string
  const list = getListOfResultTypesInCategory(category, false)

  hint = $t('search_engine.shows') + ' ' + (list.length === 1 ? $t('search_engine.this_type') : $t('search_engine.these_types')) + ' '
  for (let i = 0; i < list.length; i++) {
    hint += TypeInfo[list[i]].title
    if (i < list.length - 1) {
      hint += ', '
    }
  }

  return hint
}

// ********* THIS FUNCTION SIMULATES AN API RESPONSE - TO BE REMOVED ONCE THE API IS IMPLEMENTED *********
function simulateAPIresponse (searched : string) : SearchAheadResult {
  const response : SearchAheadResult = {}; response.data = []

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
      str_value: searched + 'Coin',
      hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
    },
    {
      chain_id: 1,
      type: 'accounts',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938'
    },
    {
      chain_id: 1,
      type: 'graffiti',
      str_value: searched + ' tutta la vita'
    },
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0x' + searched + 'a0ba39494ce839613fffba74279579260',
      str_value: 'Uniswap'
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
        num_value: Number(searched),
        hash_value: '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce'
      },
      {
        chain_id: 1,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce'
      },
      {
        chain_id: 1,
        type: 'validators_by_index',
        num_value: Number(searched),
        hash_value: '0xa525497ec3116c1310be8d73d2efd536dc0ce6bd4b0163dffddf94dad3d91d154c061b9a3bfd1b704a5ba67fc443974a'
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 1,
        type: 'tokens',
        str_value: searched,
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
      },
      {
        chain_id: 1,
        type: 'ens_addresses',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth'
      }
    )
  }
  response.data.push(
    {
      chain_id: 17000,
      type: 'contracts',
      hash_value: '0x' + searched + 'a0ba39494ce839613fffba74279579260',
      str_value: 'Uniswap'
    },
    {
      chain_id: 17000,
      type: 'validators_by_withdrawal_address',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A5357Ca938'
    },
    {
      chain_id: 42161,
      type: 'contracts',
      hash_value: '0x' + searched + '00000000000000000000000000CAFFE',
      str_value: 'Tormato Cash'
    },
    {
      chain_id: 42161,
      type: 'transactions',
      hash_value: '0x' + searched + 'a297ab886723ecfbc2cefab2ba385792058b344fbbc1f1e0a1139b2'
    },
    {
      chain_id: 8453,
      type: 'accounts',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa'
    },
    {
      chain_id: 8453,
      type: 'validators_by_deposit_address',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa'
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
        num_value: Number(searched),
        hash_value: '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca'
      },
      {
        chain_id: 8453,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca'
      },
      {
        chain_id: 8453,
        type: 'validators_by_index',
        num_value: Number(searched),
        hash_value: '0x99f9ec412465e15243a5996205928ef1461fd4ef6b6a0c642748c6f85de72c801751facda0c96454a8c2ad3bd19f91ee'
      },
      {
        chain_id: 100,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'slots',
        num_value: Number(searched),
        hash_value: '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81'
      },
      {
        chain_id: 100,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81'
      },
      {
        chain_id: 100,
        type: 'validators_by_index',
        num_value: Number(searched),
        hash_value: '0x85e5ac15a728a2bf0b0b4f22312dad780d4e27856e30997ee11f73d74d86682800046a86a01d134dbdf171326cd7cc54'
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'USD',
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
      },
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'Plus',
        hash_value: '0x0701BF988309bf45a6771afaa6B8802Ba3E24090'
      },
      {
        chain_id: 100,
        type: 'tokens',
        str_value: searched + ' Coin',
        hash_value: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'
      },
      {
        chain_id: 100,
        type: 'ens_addresses',
        str_value: searched + 'hallo.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 100,
        type: 'ens_addresses',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + 'hallo.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + 'hallo.eth'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth'
      }
    )
  }

  return response
}

// *** END OF THE FUNCTION TO BE REMOVED WHEN THE API IS IMPLEMENTED ***
</script>

<template>
  <div class="whole-engine" :class="barStyle">
    <div id="input-and-button" ref="inputFieldAndButton">
      <InputText
        id="input-field"
        v-model="inputted"
        :class="barStyle"
        type="text"
        :placeholder="$t('search_engine.placeholder')"
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
    <div v-if="showDropDown" id="drop-down" ref="dropDown">
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
            @change="networkFilterHasChanged(); refreshDropDown()"
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
                  @change="categoryFilterHasChanged(); refreshDropDown()"
                >
                <span class="face">
                  {{ CategoryInfo[filter as Category].filterLabel }}
                </span>
              </label>
            </BcTooltip>
          </span>
        </div>
      </div>
      <div v-if="inputted.length === 0" class="output-area">
        <div class="info center">
          {{ $t('search_engine.help') }}
        </div>
      </div>
      <div v-else-if="waitingForSearchResults" class="output-area">
        <div v-if="numberOfApiCallsWithoutResponse < 3" class="info center">
          {{ $t('search_engine.searching') }}
          <BcLoadingSpinner :loading="true" size="small" alignment="default" />
        </div>
        <div v-else class="info center">
          {{ $t('search_engine.something_wrong') }}
          <BcErrorIcon style="position:relative; top:2px; height:14px" />
          <br>
          {{ $t('search_engine.try_again') }}
        </div>
      </div>
      <div v-else-if="populateDropDown" class="output-area">
        <div v-for="network of results.organized.in.networks" :key="network.chainId" class="network-container">
          <div v-for="typ of network.types" :key="typ.type" class="type-container">
            <div
              v-for="(suggestion, i) of typ.suggestion"
              :key="i"
              class="single-result"
              :class="barStyle"
              @click="userClickedProposal(network.chainId, typ.type, suggestion.columns[suggestion.queryParam])"
            >
              <span v-if="network.chainId !== ChainIDs.Any" class="columns-icons">
                <IconTypeIcons :type="typ.type" class="type-icon not-alone" />
                <IconNetworkIcons :chain-id="network.chainId" class="network-icon" />
              </span>
              <span v-else class="columns-icons">
                <IconTypeIcons :type="typ.type" class="type-icon alone" />
              </span>
              <span class="columns-0">
                {{ suggestion.columns[0] }}
              </span>
              <span class="columns-1and2">
                <span v-if="suggestion.columns[1] !== ''" class="columns-1">
                  {{ suggestion.columns[1] }}
                </span>
                <span class="columns-2">
                  {{ suggestion.columns[2] }}
                </span>
              </span>
              <span class="columns-category">
                <span class="category-label">
                  {{ CategoryInfo[TypeInfo[typ.type].category].filterLabel }}
                </span>
              </span>
            </div>
          </div>
        </div>
        <div v-if="results.organized.howManyResultsIn == 0" class="info center">
          {{ $t('search_engine.no_result_matches') }}
          {{ results.organized.howManyResultsOut > 0 ? $t('search_engine.your_filters') : $t('search_engine.your_input') }}
        </div>
        <div v-if="results.organized.howManyResultsOut > 0" class="info bottom">
          {{ (results.organized.howManyResultsIn == 0 ? ' (' : '+') + String(results.organized.howManyResultsOut) }}
          {{ (results.organized.howManyResultsOut == 1 ? $t('search_engine.result_hidden') : $t('search_engine.results_hidden')) +
            (results.organized.howManyResultsIn == 0 ? ')' : ' '+$t('search_engine.by_your_filters')) }}
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
  &.discreet {
    @media (min-width: 600px) {
      // large screen
      width: 460px;
    }
    @media (max-width: 600px) {
      // mobile
      width: 380px;
    }
  }
  &.gaudy {
    @media (min-width: 600px) {
      // large screen
      width: 735px;
    }
    @media (max-width: 600px) {
      // mobile
      width: 380px;
    }
  }
}

#input-and-button {
  display: flex;

  #input-field {
    display: flex;
    flex-grow: 1;
    left: 0;
    height: v-bind(searchButtonSize);
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
    right: 0px;
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
      color: var(--input-placeholder-text-color);
    }
    &.gaudy {
      background-color: var(--button-color-active);
      font-size: 18px;
      color: var(--grey-4);
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
}

#drop-down #filter-area {
  display: flex;
  padding-top: 4px;
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
          background: var(--searchbar-filter-unselected);
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
    @media (max-width: 600px) {
      // mobile
      letter-spacing: -0.3px;  // needed to fit all the buttons in one line
    }
    .face{
      color: var(--primary-contrast-color);
      display: inline-block;
      border-radius: 10px;
      height: 17px;
      padding-top: 2.5px;
      padding-left: 8px;
      padding-right: 8px;
      text-align: center;
      margin-left: 6px;
      transition: 0.2s;
      background-color: var(--searchbar-filter-unselected);
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

#drop-down .output-area {
  display: flex;
  flex-direction: column;
  min-height: 128px;
  max-height: 270px;  // the height of the filter section is subtracted
  right: 0px;
  overflow: auto;
  @include fonts.standard_text;

  .network-container {
    display: flex;
    flex-direction: column;
    //border-bottom: 0.5px dashed var(--light-grey-3);
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
        padding-left: 2px;
        padding-right: 2px;
        padding-top: 7px;
        padding-bottom: 7px;
        @media (min-width: 600px) { // large screen
          grid-template-columns: 40px 100px auto min-content;
          &.gaudy {
            padding-left: 4px;
            padding-right: 4px;
          }
        }
        @media (max-width: 600px) { // mobile
          grid-template-columns: 40px 100px auto;
        }
        border-radius: var(--border-radius);

        &:hover {
          background-color: var(--dropdown-background-hover);
        }
        &:active {
          background-color: var(--button-color-pressed);
        }

        .columns-icons {
          position: relative;
          grid-column: 1;
          grid-row: 1;
          @media (max-width: 600px) { // mobile
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
            height: 20px;
            max-width:20px;
          }
        }
        .columns-0 {
          grid-column: 2;
          grid-row: 1;
          display: flex;
          margin-top: auto;
          margin-bottom: auto;
          overflow-wrap: anywhere;
          font-weight: var(--roboto-medium);
          padding-right: 4px;
        }
        .columns-1and2 {
          min-width: 0;
          grid-column: 3;
          grid-row: 1;
          @media (max-width: 600px) { // mobile
            grid-row-end: span 2;
          }
          display: flex;
          margin-top: auto;
          margin-bottom: auto;
          font-weight: var(--roboto-medium);
          .columns-1 {
            display: flex;
            overflow-wrap: break-word;
            margin-right: 0.8ch;
          }
          .columns-2 {
            display: flex;
            min-width: 0;
            overflow-wrap: anywhere;
          }
        }
        .columns-category {
          @media (min-width: 600px) { // large screen
            grid-column: 4;
            grid-row: 1;
            display: flex;
            margin-top: auto;
            margin-bottom: auto;
          }
          @media (max-width: 600px) { // mobile
            grid-column: 2;
            grid-row: 2;
          }
          .category-label {
            @media (min-width: 600px) { // large screen
              float: right;
              margin-left: 8px;
            }
            color: var(--drop-down-text-discreet);
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
      margin-top: auto;
    }
    &.center {
      margin-bottom: auto;
      margin-top: auto;
    }
  }
}
</style>
