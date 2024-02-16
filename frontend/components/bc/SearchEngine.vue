<script setup lang="ts">
import { Categories, ResultTypes, TypeInfo, organizeAPIinfo, type SearchAheadResults, type OrganizedResults } from '~/types/search'
import { ChainIDs, ChainInfo, getListOfChainIDs } from '~/types/networks'
const { t: $t } = useI18n()

const props = defineProps({ searchable: { type: String, required: true }, width: { type: String, required: true }, height: { type: String, required: true } })
const searchable = props.searchable as Categories
const emit = defineEmits(['enter', 'select'])

const engineWidth = props.width + 'px'
const inputWidth = String(Number(props.width) - 10) + 'px'
const dropDownWidth = String(Number(props.width) - 10) + 'px'
const inputHeight = props.height + 'px'

const PeriodOfDropDownUpdates = 2000
const APIcallTimeout = 1500 // should not exceed PeriodOfDropDownUpdates

const waitingForSearchResults = ref(false)
const showDropDown = ref(false)
const populateDropDown = ref(true)
const inputField = ref('')
let lastKnownInput = ''
let isMouseOverEngine = false

const results = {
  raw: { data: [] } as SearchAheadResults, // response of the API, without structure nor order
  organized: {
    in: { networks: [] } as OrganizedResults, // filtered results, organized
    out: { networks: [] } as OrganizedResults // filtered out results, organized
  }
}

interface UserFilters {
  ['network']: string,
  [Categories.Tokens]: 'y'|'n',
  [Categories.NFTs]: 'y'|'n',
  [Categories.Protocol]: 'y'|'n',
  [Categories.Addresses]: 'y'|'n',
  [Categories.Validators]: 'y'|'n'
}
const userFilters = ref<UserFilters>({
  network: 'all',
  [Categories.Tokens]: 'n',
  [Categories.NFTs]: 'n',
  [Categories.Protocol]: 'n',
  [Categories.Addresses]: 'n',
  [Categories.Validators]: 'n'
})
const networkButtonColor = ref('var(--light-grey-3)')
function setNetworkButtonColor () {
  networkButtonColor.value = (userFilters.value.network === 'all') ? 'var(--light-grey-3)' : 'var(--primary-color)'
}

function cleanUp () {
  lastKnownInput = ''
  inputField.value = ''
  waitingForSearchResults.value = false
  showDropDown.value = false
  populateDropDown.value = true
  isMouseOverEngine = false
  results.raw = { data: [] }
  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }
}

// In the V1, the server received a request between 1.5 and 3.5 seconds after the user inputted something, depending on the length of the input.
// Therefore, the average delay was ~2.5 s for the user as well as for the server. Most of the time the delay was shorter because the 3.5 s delay
// was only for entries of size 1.
// This less-than-2.5s-on-average delay arised from a Timeout Timer.
// For the V2, I propose to work with a 2-second Interval Timer because:
// - it makes sure that the server does not receive a request more often than every 2 s (equivalent to V1),
// - while offering the user an average waiting time of 1 second through the magic of statistics (better than V1).
setInterval(() => {
  if (waitingForSearchResults.value) {
    if (searchAhead(inputField.value, searchable)) {
      filterAndOrganizeResults()
      waitingForSearchResults.value = false
    }
  }
},
PeriodOfDropDownUpdates
)

function inputMightHaveChanged () {
  if (inputField.value === lastKnownInput) {
    return
  }
  lastKnownInput = inputField.value
  if (inputField.value.length === 0) {
    cleanUp()
  } else {
    waitingForSearchResults.value = true
    showDropDown.value = true
  }
}

function userPressedEnter () {
  if (inputField.value.length === 0) {
    return
  }
  if (waitingForSearchResults.value) {
    if (!searchAhead(inputField.value, searchable)) {
      return
    }
  }
  if (areResultsEmpty(false)) {
    // false in the test means that we will pick a filtered-out result if the drop down does not show anything although there are results
    return
  }
  // picks a relevant search-ahead result
  // **** TO BE CHANGED ONCE THE NETWORK DROPDOWN IS IMPLEMENTED ****
  const userPreferredNetwork = ChainIDs.Ethereum
  // ****************************************************************
  filterAndOrganizeResults()
  let defaultNetwork = results.organized.in.networks[0]
  for (const network of results.organized.in.networks) {
    if (network.chainId === userPreferredNetwork) {
      defaultNetwork = network
      break
    }
  }
  const defaultType = defaultNetwork.types[0]
  // cleans up and calls back user's function with the first result
  cleanUp()
  emit('enter', defaultType.found[0].main, defaultType.type, defaultNetwork.chainId)
}

function userClickedProposal (chain : ChainIDs, type : ResultTypes, found: string) {
  // cleans up and calls back user's function
  cleanUp()
  emit('select', found, type, chain)
}

// returns false if the API could not be reached or if it had a problem
// returns true otherwise (so also true when no result matches the input)
function searchAhead (input : string, searchable : Categories) : boolean {
  let error = false

  // ********* SIMULATES AN API RESPONSE - TO BE REMOVED ONCE THE API IS IMPLEMENTED *********
  if (searchable as string !== '-- to be removed --') {
    results.raw = simulateAPIresponse(input)
  } else { // *** END OF STUFF TO REMOVE ***
    fetch('/api/2/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input, searchable }),
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
  return !error
}

// Fills `results.organized` by categorizing and filtering the disorganized data of the API.
function filterAndOrganizeResults () {
  // determining whether any filter button is activated
  let areAllButtonsOff = true
  for (const k of Object.keys(userFilters.value)) {
    if (userFilters.value[k as keyof UserFilters] === 'y') {
      areAllButtonsOff = false
      break
    }
  }

  results.organized.in = { networks: [] }
  results.organized.out = { networks: [] }

  if (results.raw.data === undefined) {
    return
  }
  for (const finding of results.raw.data) {
    const chainId = finding.chain_id as ChainIDs
    const type = finding.type as ResultTypes

    // getting organized information from the finding
    const toBeAdded = organizeAPIinfo(finding)
    if (toBeAdded.main === '') {
      continue
    }
    // determining whether the finding is filtered in or out, pointing `place` to the corresponding object
    let place : OrganizedResults
    if ((userFilters.value.network === String(chainId) || userFilters.value.network === 'all') &&
        (userFilters.value[TypeInfo[type].category as keyof UserFilters] === 'y' || areAllButtonsOff /* if all filters are inactive, we default to showing everything */)) {
      place = results.organized.in
    } else {
      place = results.organized.out
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
    place.networks[existingNetwork].types[existingType].found.push(toBeAdded)
  }
}

function refreshDropDown () {
  populateDropDown.value = false
  filterAndOrganizeResults()
  populateDropDown.value = true // this triggers Vue to refresh the list of results
}

function areResultsEmpty (considerFilteredOutResults : boolean) : boolean {
  return results.organized.in.networks.length === 0 && (!considerFilteredOutResults || results.organized.out.networks.length === 0)
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
      str_value: searched
    },
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
      }
    )
  }
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 17000,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 17000,
        type: 'slots',
        num_value: Number(searched)
      },
      {
        chain_id: 17000,
        type: 'blocks',
        num_value: Number(searched)
      },
      {
        chain_id: 17000,
        type: 'validators_by_index',
        num_value: Number(searched)
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 17000,
        type: 'tokens',
        str_value: searched + ' Coin'
      },
      {
        chain_id: 17000,
        type: 'ens_names',
        str_value: searched + 'hallo.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 17000,
        type: 'ens_names',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 17000,
        type: 'ens_overview',
        str_value: searched + 'hallo.eth'
      },
      {
        chain_id: 17000,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth'
      },
      {
        chain_id: 17000,
        type: 'count_validators_by_withdrawal_ens_name',
        str_value: searched + 'hallo.eth',
        num_value: 2
      },
      {
        chain_id: 17000,
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
  <label id="whole-engine">
    <input
      id="input-field"
      v-model="inputField"
      type="text"
      @keyup="(e) => {if (e.key === 'Enter') {userPressedEnter()} else {inputMightHaveChanged()}}"
      @blur="showDropDown = isMouseOverEngine"
    >
    <div
      v-if="showDropDown"
      id="drop-down"
      @mouseenter="isMouseOverEngine = true"
      @mouseleave="isMouseOverEngine = false"
    >
      <div v-if="waitingForSearchResults">
        {{ $t('search_engine.searching') }}
      </div>
      <div v-else-if="areResultsEmpty(true)">
        {{ $t('search_engine.no_result') }}
      </div>
      <div v-else>
        <div id="filter-bar">
          <label><select
            id="filter-list"
            v-model="userFilters['network']"
            class="filter-button"
            @change="setNetworkButtonColor(); refreshDropDown()"
          >
            <option value="all">All networks</option>
            <option v-for="chain in getListOfChainIDs()" :key="chain" :value="String(chain)">
              {{ ChainInfo[chain].name }}
            </option>
          </select>
          </label>
          <label>
            <input
              v-model="userFilters[Categories.Protocol]"
              class="filter-cb"
              true-value="y"
              false-value="n"
              type="checkbox"
              @change="refreshDropDown()"
            >
            <span class="filter-button">Protocol</span>
          </label>
          <label>
            <input
              v-model="userFilters[Categories.Addresses]"
              class="filter-cb"
              true-value="y"
              false-value="n"
              type="checkbox"
              @change="refreshDropDown()"
            >
            <span class="filter-button">Addresses</span>
          </label>
          <label>
            <input
              v-model="userFilters[Categories.Tokens]"
              class="filter-cb"
              true-value="y"
              false-value="n"
              type="checkbox"
              @change="refreshDropDown()"
            >
            <span class="filter-button">Tokens</span>
          </label>
          <label>
            <input
              v-model="userFilters[Categories.NFTs]"
              class="filter-cb"
              true-value="y"
              false-value="n"
              type="checkbox"
              @change="refreshDropDown()"
            >
            <span class="filter-button">NFTs</span>
          </label>
          <label>
            <input
              v-model="userFilters[Categories.Validators]"
              class="filter-cb"
              true-value="y"
              false-value="n"
              type="checkbox"
              @change="refreshDropDown()"
            >
            <span class="filter-button">Validators</span>
          </label>
        </div>
        <span v-if="populateDropDown">
          <div v-for="network in results.organized.in.networks" :key="network.chainId" class="network-container">
            <div class="network-title">
              <h2>{{ ChainInfo[network.chainId].name }}</h2>
            </div>
            <div v-for="types in network.types" :key="types.type" class="type-container">
              <div class="type-title">
                <h3>{{ TypeInfo[types.type].title }}</h3>
              </div>
              <div v-for="(found, i) in types.found" :key="i" class="single-result" @click="userClickedProposal(network.chainId, types.type, found.main)">
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
        </span>
      </div>
    </div>
  </label>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

#whole-engine {
  width: v-bind(engineWidth);
}

#input-field {
  display: block;
  width: v-bind(inputWidth);
  height: v-bind(inputHeight);
}

#drop-down {
  @include main.container;
  position: absolute;
  z-index: 256;
  overflow: auto;
  max-height: 66vh;
  width: v-bind(dropDownWidth);
  padding: 4px;
}

.network-container {
  margin-bottom: 24px;
}

.network-title {
  background-color: #b0b0b0;
  padding-left: 4px;
}

.type-title {

}
.type-container {
  border-bottom: 0.5px dashed var(--light-grey-3);
  padding: 4px;
}

.single-result {
  cursor: pointer;
}

h2 {
  margin: 0;
}

h3 {
  margin: 0;
}

#filter-bar {
  padding-top: 4px;
  padding-bottom: 8px;
}

#filter-list {
  background: v-bind(networkButtonColor);
}

.filter-cb {
  display: none;
  width: 0;
  height: 0;
}

.filter-button {
  display: inline-block;
  border-radius: 6px;
  background-color: var(--light-grey-3);
  padding: 2px;
  width: 80px;
  text-align: center;
  margin-right: 6px;
  transition: 0.2s;
}
.filter-cb:checked + .filter-button {
  background-color: var(--primary-color);
}
</style>
