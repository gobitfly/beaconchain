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

export type BlockchainSearchFilters = 'address' | 'block' | 'epoch' | 'slot' | 'token' | 'transaction' | 'validator'

const emit = defineEmits<{
  (e: 'search', input: string): void,
  (e: 'click:example', type: BlockchainSearchFilters): void,
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
    label: $t('products.landing_page.search.types.blocks'),
    value: 'block',
  },
  {
    label: $t('products.landing_page.search.types.epochs'),
    value: 'epoch',
  },
  {
    label: $t('products.landing_page.search.types.slots'),
    value: 'slot',
  },
  {
    label: $t('products.landing_page.search.types.tokens'),
    value: 'token',
  },
  {
    label: $t('products.landing_page.search.types.transactions'),
    value: 'transaction',
  },
  {
    label: $t('products.landing_page.search.types.validators_indices'),
    value: 'validator_by_index',
  },
]

const handleSearch = (input: string) => {
  isHistoryVisible.value = false
  emit('search', input)
}

const handleTypeFilterChange = () => {
  // When all filter chips are selected, we also want to include all available types
  // that are not included in the chips
  if (chips.every(chip => searchParams.value.types.includes(chip.value))) {
    searchParams.value.types = typeFilters
  }

  handleSearch(searchParams.value.input)
}
const history = useLocalStorage<string[]>('bc-search-history-product-landing', [])
// using localHistory instead of history directly to avoid
// that the search history in the UI is updated before navigating away
const localHistory = ref<InternalPostSearchResponseWithChainId['data']>(history.value.map(item => JSON.parse(item)))
const hasHistory = computed(() => !!localHistory.value.length)
const hasResults = computed(() => results !== undefined)
const hasInput = computed(() => searchParams.value.input.length > 0)

const isHistoryVisible = ref<boolean>(!hasResults.value && hasHistory.value)
const resultsOrHistory = computed(() => {
  if ((!hasInput.value && hasHistory.value) || isHistoryVisible.value) {
    return localHistory.value
  }
  return results
})
const toggleHistory = () => {
  isHistoryVisible.value = !isHistoryVisible.value
  localHistory.value = history.value.map(item => JSON.parse(item))
}
const handleSelect = (searchResult: InternalPostSearchResponseWithChainId['data'][number]) => {
  const currentEntry = JSON.stringify(searchResult)
  if (history.value.length >= 10) {
    history.value.pop()
  }
  history.value = history.value.filter(entry => entry !== currentEntry)
  history.value.unshift(currentEntry)
}
watch(hasResults, () => {
  if (!hasHistory.value) return
  if (hasResults.value) return
  isHistoryVisible.value = true
})
const searchInput = useTemplateRef<ComponentPublicInstance | null>('searchInput')
const handleClickExample = (type: BlockchainSearchFilters) => {
  emit('click:example', type)
  isHistoryVisible.value = false
  const input = searchInput.value?.$el.querySelector('input')
  input?.focus()
}
</script>

<template>
  <BaseSearchInput
    ref="searchInput"
    v-model="searchParams.input"
    :is-loading="isHistoryVisible ? false : isLoading"
    :has-error="isHistoryVisible ? false : hasError"
    :label="$t('products.landing_page.search.input_label')"
    :placeholder="$t('products.landing_page.search.input_placeholder')"
    :group-by="'type'"
    :results="resultsOrHistory"
    @search="handleSearch"
  >
    <template #search-examples>
      <div class="flex items-center">
        <section class="flex gap-lg">
          <div class="py-xs px-md border-gray-400 font-semibold text-gray-400">
            {{ $t('products.landing_page.search.examples.title') }}
          </div>
          <BaseChip
            :is-selected="false"
            icon="switch-horizontal"
            :aria-label="$t('products.landing_page.search.examples.transaction')"
            @click="emit('click:example', 'transaction')"
          >
            {{ $t('products.landing_page.search.examples.tx') }}
          </BaseChip>
          <BaseChip
            :is-selected="false"
            icon="hash"
            @click="emit('click:example', 'address')"
          >
            {{ $t('products.landing_page.search.examples.address') }}
          </BaseChip>
          <BaseChip
            :is-selected="false"
            icon="stack-2"
            @click="handleClickExample('validator')"
          >
            {{ $t('products.landing_page.search.examples.validator') }}
          </BaseChip>
          <BaseChip
            :is-selected="false"
            icon="hexagon"
            @click="emit('click:example', 'token')"
          >
            {{ $t('products.landing_page.search.examples.token') }}
          </BaseChip>
        </section>
      </div>
    </template>
    <template #dropdown-fixed-header="{ idSearchInput }">
      <div
        class="min-h-fit overflow-x-auto overscroll-contain flex gap-md items-center px-2xl py-lg"
        @keydown.enter.stop
      >
        <BaseButtonIcon
          v-if="hasHistory"
          :aria-controls="idSearchInput"
          :is-disabled="!hasResults"
          role="switch"
          screenreader-text="products.landing_page.search.history.action.toggle_history"
          name="history"
          :aria-checked="`${isHistoryVisible}`"
          variant="secondary"
          @click="toggleHistory"
        />
        <BaseChipGroup
          v-if="!isHistoryVisible"
          v-model="searchParams.types"
          :aria-controls="idSearchInput"
          :items="chips"
          :aria-label="$t('products.landing_page.search.filter_aria_label')"
          @update:model-value="handleTypeFilterChange"
        />
        <span v-else>
          {{ $t('products.landing_page.search.history.recent') }}
        </span>
      </div>
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
      <BlockchainSearchResultItem
        :result
        @click="handleSelect(result)"
      />
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
