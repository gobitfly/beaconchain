<script setup lang="ts">
import { warn } from 'vue'
import { Searchable, ResultTypes, TypeInfo, organizeAPIinfo, type SearchAheadResults, type OrganizedResults } from '~/types/search'
import { ChainIDs, ChainInfo } from '~/types/networks'

const props = defineProps({ searchable: { type: String, required: true } })
const searchable = props.searchable as Searchable
const emit = defineEmits(['enter', 'select'])

const PeriodOfDropDownUpdates = 2000
const APIcallTimeout = 1500 // should not exceed PeriodOfDropDownUpdates

const newCharGivenSinceSearch = ref(false)
const showDropDown = ref(false)
const inputField = ref('')
let organizedResults : OrganizedResults = { networks: [] }
let lastKnownInput = ''

function cleanUp () {
  lastKnownInput = ''
  inputField.value = ''
  newCharGivenSinceSearch.value = false
  showDropDown.value = false
  organizedResults = { networks: [] }
}

// In the V1, the server received a request between 1.5 and 3.5 seconds after the user inputted something, depending on the length of the input.
// Therefore, the average delay was ~2.5 s for the user as well as for the server. Most of the time the delay was shorter because the 3.5 s delay
// was only for entries of size 1.
// This less-than-2.5s-on-average delay arised from a Timeout Timer.
// For the V2, I propose to work with a 2-second Interval Timer because:
// - it makes sure that the server does not receive a request more often than every 2 s (equivalent to V1),
// - while offering the user an average waiting time of 1 second through the magic of statistics (better than V1).
setInterval(() => {
  if (newCharGivenSinceSearch.value) {
    newCharGivenSinceSearch.value = !searchAhead(inputField.value, searchable)
    // this assignement ensures that the API will be called again in 2s if searchAhead fails for technical reasons
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
    newCharGivenSinceSearch.value = true
    showDropDown.value = true
  }
}

function userPressedEnter () {
  if (inputField.value.length === 0) {
    return
  }
  if (newCharGivenSinceSearch.value) {
    if (!searchAhead(inputField.value, searchable)) {
      return
    }
  }
  if (isOrganizedResultsEmpty()) {
    return
  }
  // picks a relevant search-ahead result
  // **** TO BE CHANGED ONCE THE NETWORK DROPDOWN IS IMPLEMENTED ****
  const userPreferredNetwork = ChainIDs.Ethereum
  // ****************************************************************
  let defaultNetwork = organizedResults.networks[0]
  for (const network of organizedResults.networks) {
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
function searchAhead (input : string, searchable : Searchable) : boolean {
  let foundAhead : SearchAheadResults = { data: [], error: '' }

  // ********* SIMULATES AN API RESPONSE - TO BE REMOVED ONCE THE API IS IMPLEMENTED *********
  if (searchable as string !== 'please remove this condition and the following code') {
    if (Math.random() > 1 / 3) {
      foundAhead = {
        data: [
          {
            chain_id: 1,
            type: 'tokens',
            str_value: input
          },
          {
            chain_id: 1,
            type: 'tokens',
            str_value: input + 'Coin'
          },
          {
            chain_id: 1,
            type: 'addresses',
            hash_value: '0x' + input + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938'
          },
          {
            chain_id: 1,
            type: 'ens_names',
            str_value: input + '.bitfly.eth',
            hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
          },
          {
            chain_id: 1,
            type: 'ens_overview',
            str_value: input + '.bitfly.eth'
          },
          {
            chain_id: 1,
            type: 'graffiti',
            str_value: input + ' tutta la vita'
          },
          {
            chain_id: 1,
            type: 'count_validators_by_withdrawal_ens_name',
            str_value: input + '.bitfly.eth',
            num_value: 7
          },
          {
            chain_id: 17000,
            type: 'addresses',
            hash_value: '0x' + input + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938'
          },
          {
            chain_id: 17000,
            type: 'count_validators_by_withdrawal_address',
            hash_value: '0x' + input + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938',
            num_value: 11
          },
          {
            chain_id: 42161,
            type: 'addresses',
            hash_value: '0x' + input + '0000000000000000000000000000CAFFE'
          },
          {
            chain_id: 42161,
            type: 'transactions',
            hash_value: '0x' + input + 'a297ab886723ecfbc2cefab2ba385792058b344fbbc1f1e0a1139b2'
          },
          {
            chain_id: 8453,
            type: 'tokens',
            str_value: input + 'USD'
          },
          {
            chain_id: 8453,
            type: 'tokens',
            str_value: input + '42'
          },
          {
            chain_id: 8453,
            type: 'tokens',
            str_value: input + 'Plus'
          },
          {
            chain_id: 8453,
            type: 'addresses',
            hash_value: '0x' + input + '00b29F2d2FaDE0a7E3AAaaAAa'
          },
          {
            chain_id: 8453,
            type: 'count_validators_by_deposit_address',
            hash_value: '0x' + input + '00b29F2d2FaDE0a7E3AAaaAAa',
            num_value: 150
          },
          {
            chain_id: 17000,
            type: 'tokens',
            str_value: input + ' Coin'
          },
          {
            chain_id: 17000,
            type: 'ens_names',
            str_value: input + 'hallo.eth',
            hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
          },
          {
            chain_id: 17000,
            type: 'ens_names',
            str_value: input + '.bitfly.eth',
            hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
          },
          {
            chain_id: 17000,
            type: 'ens_overview',
            str_value: input + 'hallo.eth'
          },
          {
            chain_id: 17000,
            type: 'ens_overview',
            str_value: input + '.bitfly.eth'
          },
          {
            chain_id: 17000,
            type: 'count_validators_by_withdrawal_ens_name',
            str_value: input + 'hallo.eth',
            num_value: 2
          },
          {
            chain_id: 17000,
            type: 'count_validators_by_withdrawal_ens_name',
            str_value: input + '.bitfly.eth',
            num_value: 150
          }
        ]
      }
    }
    // *** END OF STUFF TO BE REMOVED WHEN THE API IS IMPLEMENTED ***
  } else {
    fetch('/api/2/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input, searchable }),
      signal: AbortSignal.timeout(APIcallTimeout)
    }).then((received) => {
      if (received.ok && received.status < 400) {
        received.json().then((object) => {
          foundAhead = object
        })
      } else {
        return false
      }
    }).catch(() => {
      return false
    })
    if (foundAhead === undefined || foundAhead.error !== undefined) {
      return false
    }
  }

  // now we take the disorganized data of the API and fill `organizedResults`, which will be easy to iterate over when populating the drop-down
  organizedResults = { networks: [] }
  if (foundAhead.data !== undefined && foundAhead.data.length > 0) {
    for (const finding of foundAhead.data) {
      const toBeAdded = organizeAPIinfo(finding)
      if (toBeAdded.main === '') {
        continue
      }
      // Picking from `organizedResults` the network that the finding belongs to. Creates the network if needed.
      let existingNetwork = organizedResults.networks.findIndex(nwElem => nwElem.chainId === finding.chain_id as ChainIDs)
      if (existingNetwork < 0) {
        existingNetwork = -1 + organizedResults.networks.push({
          chainId: finding.chain_id as ChainIDs,
          types: []
        })
      }
      // Picking from the network the type group that the finding belongs to. Creates the type group if needed.
      let existingType = organizedResults.networks[existingNetwork].types.findIndex(tyElem => tyElem.type === finding.type as ResultTypes)
      if (existingType < 0) {
        existingType = -1 + organizedResults.networks[existingNetwork].types.push({
          type: finding.type as ResultTypes,
          found: []
        })
      }
      // now we can insert the finding at the right place in `organizedResults`
      organizedResults.networks[existingNetwork].types[existingType].found.push(toBeAdded)
    }
  }

  return true
}

function isOrganizedResultsEmpty () {
  return organizedResults.networks.length === 0
}
</script>

<template>
  <div>
    <label><input
      id="input-field"
      v-model="inputField"
      type="text"
      @keyup="(e) => {if (e.key === 'Enter') {userPressedEnter()} else {inputMightHaveChanged()}}"
    ></label>
    <div v-if="showDropDown" id="drop-down">
      <div v-if="newCharGivenSinceSearch">
        Searching...
      </div>
      <div v-else-if="isOrganizedResultsEmpty()">
        No result
      </div>
      <div v-else>
        <div v-for="network in organizedResults.networks" :key="network.chainId" class="network-frame">
          <div><h2>{{ ChainInfo[network.chainId].name }}</h2></div>
          <div v-for="types in network.types" :key="types.type" class="results-of-one-type">
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
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

#input-field {

}

#drop-down {
  @include main.container;
  position: absolute;
  z-index: 100;
}

.network-frame {

}

.results-of-one-type {

}

.type-title {

}

.single-result {
  cursor: pointer;
}
</style>
