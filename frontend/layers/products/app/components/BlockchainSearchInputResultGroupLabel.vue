<script setup lang="ts">
import type { InternalPostSearchResponse } from '~/types/api/search'

export type SearchResultGroupType = InternalPostSearchResponse['data'][number]['type']

defineProps<{
  groupType: SearchResultGroupType,
}>()
</script>

<template>
  <RkComboboxLabel class="px-2xl py-md  dark:text-gray-400">
    <span v-if="groupType === 'address'"> Addresses</span>
    <span v-else-if="groupType === 'block'"> Blocks</span>
    <span v-else-if="groupType === 'ens_name'"> ENS Names</span>
    <span v-else-if="groupType === 'epoch'"> Epochs</span>
    <span v-else-if="groupType === 'slot'"> Slots</span>
    <span v-else-if="groupType === 'token'"> Tokens</span>
    <span v-else-if="groupType === 'transaction'"> Tx Hashes</span>
    <span v-else-if="groupType === 'validator'"> Validators</span>
    <span v-else-if="groupType === 'validators_by_deposit_address'"> Validators by Deposit Address</span>
    <span v-else-if="groupType === 'validators_by_withdrawal_credential'"> Validators by Withdrawal Credential</span>
  </RkComboboxLabel>
</template>
