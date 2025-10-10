<script setup lang="ts">
import type { PostSearchRequest } from '~/types/api/search'
import type { InternalPostSearchResponseWithChainId } from '~/layers/products/server/api/bff/search'

export type BlockchainSearchParams = Omit<PostSearchRequest, 'networks' | 'types'> & { networks: ChainId[], types: BlockchainSearchTypes }
export type BlockchainSearchTypes = Exclude<NonNullable<PostSearchRequest['types']>[number], 'validator_list'>[]

const {
  hasError,
  isLoading,
  results,
  typeFilters,
} = defineProps<{
  hasError?: boolean,
  isLoading: boolean,
  results?: InternalPostSearchResponseWithChainId['data'],
  typeFilters: BlockchainSearchParams['types'],
}>()

const { t: $t } = useTranslation()

const emit = defineEmits<{
  (e: 'search'): void,
}>()

const searchParams = defineModel<BlockchainSearchParams>({
  required: true,
})

const chips: { label: string, value: BlockchainSearchParams['types'][number] }[] = [
  {
    label: $t('products.landing_page.search.types.addresses'),
    value: 'address',
  },
  {
    label: $t('products.landing_page.search.types.transactions'),
    value: 'transaction',
  },
  {
    label: $t('products.landing_page.search.types.validators_indices'),
    value: 'validator_by_index',
  },
  {
    label: $t('products.landing_page.search.types.blocks'),
    value: 'block',
  },
  {
    label: $t('products.landing_page.search.types.tokens'),
    value: 'token',
  },
]

const handleSearch = () => {
  emit('search')
}

const handleTypeFilterChange = () => {
  // When all filter chips are selected, we also want to include all available types
  // that are not included in the chips
  if (chips.every(chip => searchParams.value.types.includes(chip.value))) {
    searchParams.value.types = typeFilters
  }

  handleSearch()
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
    <template #dropdown-fixed-header>
      <BaseChipGroup
        v-model="searchParams.types"
        :items="chips"
        class="overflow-x-auto overscroll-contain min-h-fit"
        :aria-label="$t('products.landing_page.search.filter_aria_label')"
        @update:model-value="handleTypeFilterChange"
      />
      <hr class="mx-2xl text-gray-600">
    </template>

    <template #result-group-label="{ label }">
      <span v-if="label === 'address'">{{ $t('products.landing_page.search.types.addresses') }}</span>
      <span v-else-if="label === 'block'">{{ $t('products.landing_page.search.types.blocks') }}</span>
      <span v-else-if="label === 'ens_name'">{{ $t('products.landing_page.search.types.ens_names') }}</span>
      <span v-else-if="label === 'epoch'">{{ $t('products.landing_page.search.types.epochs') }}</span>
      <span v-else-if="label === 'slot'">{{ $t('products.landing_page.search.types.slots') }}</span>
      <span v-else-if="label === 'token'">{{ $t('products.landing_page.search.types.tokens') }}</span>
      <span v-else-if="label === 'transaction'">{{ $t('products.landing_page.search.types.transactions') }}</span>
      <span v-else-if="label === 'validator'">{{ $t('products.landing_page.search.types.validators_indices') }}</span>
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
