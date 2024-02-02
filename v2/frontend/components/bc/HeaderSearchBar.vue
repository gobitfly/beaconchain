<script setup lang="ts">
import { warn } from 'vue'
import { SearchTypes } from '~/types/searchtypes'

function userPressedEnter (searched : string, type : SearchTypes) {
  navigateTo({
    path: '/search',
    query: {
      page: 1,
      sort: 'asc'
    }
  })
  warn(searched, ' ', type)
}
function userClickedTheDropdown (searched : string, type : SearchTypes) {
  warn(searched, ' ', type)
}
</script>

<template>
  <BcSearchEngine
    :searchable="[
      SearchTypes.slots, SearchTypes.blocks, SearchTypes.graffiti, SearchTypes.transactions, SearchTypes.epochs,
      SearchTypes.validators, SearchTypes.addresses, SearchTypes.validators_by_pubkey, SearchTypes.count_indexed_validators_by_deposit_address,
      SearchTypes.count_indexed_validators_by_withdrawal_credential, SearchTypes.ens, SearchTypes.network
    ]"
    @enter="userPressedEnter"
    @select="userClickedTheDropdown"
  />
</template>
