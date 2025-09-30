<script setup lang="ts">
import type { PostSearchRequest } from '~/types/api/search'
import type { InternalPostSearchResponseWithChainId } from '~/layers/products/server/api/bff/search'

export type SearchParams = Omit<PostSearchRequest, 'networks'> & { networks: ChainId[] }

const {
  hasError,
  isLoading,
  results,
} = defineProps<{
  hasError?: boolean,
  isLoading: boolean,
  results?: InternalPostSearchResponseWithChainId['data'],
}>()

const { t: $t } = useTranslation()

const emit = defineEmits<{
  (e: 'search', searchParams: SearchParams): void,
}>()

const searchParams = defineModel<SearchParams>({
  required: true,
})

const handleSearch = (input: string) => {
  searchParams.value.input = input
  emit('search', searchParams.value)
}
</script>

<template>
  <BaseSearchInput
    v-model="searchParams.input"
    :is-loading
    :has-error
    :label="$t('products.landing_page.search.input_label')"
    :placeholder="$t('products.landing_page.search.input_placeholder')"
    :group-by="'type'"
    :results
    @search="handleSearch"
  >
    <template #result-group-label="{ label }">
      <span v-if="label === 'address'">{{ $t('products.landing_page.search.types.addresses') }}</span>
      <span v-else-if="label === 'block'">{{ $t('products.landing_page.search.types.blocks') }}</span>
      <span v-else-if="label === 'ens_name'">{{ $t('products.landing_page.search.types.ens_names') }}</span>
      <span v-else-if="label === 'epoch'">{{ $t('products.landing_page.search.types.epochs') }}</span>
      <span v-else-if="label === 'slot'">{{ $t('products.landing_page.search.types.slots') }}</span>
      <span v-else-if="label === 'token'">{{ $t('products.landing_page.search.types.tokens') }}</span>
      <span v-else-if="label === 'transaction'">{{ $t('products.landing_page.search.types.transactions') }}</span>
      <span v-else-if="label === 'validator'">{{ $t('products.landing_page.search.types.validators') }}</span>
      <span v-else-if="label === 'validators_by_deposit_address'">
        {{ $t('products.landing_page.search.types.validator_deposit_addresses') }}
      </span>
      <span v-else-if="label === 'validators_by_graffiti'">
        {{ $t('products.landing_page.search.types.validator_graffiti') }}</span>
      <span v-else-if="label === 'validators_by_withdrawal_credential'">
        {{ $t('products.landing_page.search.types.validator_withdrawal_credentials') }}
      </span>
    </template>

    <template #result-item="{ result }">
      <BlockchainSearchResultItem :result />
    </template>

    <template #loading-content>
      <BlockchainSearchLoadingSkeleton />
    </template>
  </BaseSearchInput>
</template>

<style lang="scss" scoped>
.search-input::-webkit-search-cancel-button {
  display: none;
}
</style>
