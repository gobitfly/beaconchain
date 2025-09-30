<script setup lang="ts" generic="SearchResultType extends Record<string, any>">
type GroupBy = keyof SearchResultType

const {
  groupBy,
  hasError,
  isLoading,
  results,
} = defineProps<{
  groupBy?: GroupBy,
  hasError?: boolean,
  isLoading: boolean,
  label: string,
  placeholder?: string,
  results?: SearchResultType[],
}>()

const emit = defineEmits<{
  (e: 'search', input: string): void,
}>()

const { t: $t } = useTranslation()

const hasSearched = ref(false)
const input = defineModel<string>({
  required: true,
})

watchDebounced(
  input,
  async () => {
    if (input.value.length) {
      emit('search', input.value)
      hasSearched.value = true
    }
  },
  {
    immediate: false,
  },
)

const showDropdown = computed(() => isLoading || hasSearched.value)
const groupedResults = computed(() => {
  if (!results?.length) return
  if (!groupBy) return

  const groupedResults = Object.groupBy(results, result => result[groupBy] as PropertyKey)

  return Object.entries(groupedResults) as [SearchResultType[GroupBy], SearchResultType[]][]
})
</script>

<template>
  <form
    role="search"
    class="base-search-input__form p-2xl"
  >
    <RkComboboxRoot
      v-model:open="showDropdown"
      class="relative"
      ignore-filter
    >
      <RkLabel
        for="search-input"
        class="absolute bottom-2xl left-2xl dark:text-gray-400 text-sm-tight"
      >
        {{ label }}
      </RkLabel>
      <RkComboboxInput
        id="search-input"
        v-model.trim="input"
        type="search"
        auto-focus
        :aria-busy="isLoading"
        :placeholder
        class="search-input w-full text-2xl font-semibold rounded-3xl pt-3xl pr-5xl pb-6xl pl-2xl dark:bg-gray-950
        dark:focus:bg-black placeholder:dark:text-gray-500 dark:text-white border-1 dark:border-charcoal-500
        dark:focus:border-charcoal-50 dark:focus-within:outline-0"
        @update:model-value="(value) => { if (!value) hasSearched = false }"
      />

      <RkComboboxContent class="absolute dark:bg-gray-950 mt-xl rounded-xl w-full max-h-[400px] overflow-y-auto">
        <!-- SLOTS FOR HISTORY AND FILTERS GO HERE -->
        <RkComboboxViewport>
          <div class="py-lg">
            <slot
              v-if="isLoading"
              name="loading-content"
            />

            <slot
              v-else-if="hasError"
              name="error-content"
            >
              <div
                role="alert"
                class="px-2xl py-md dark:text-gray-400 "
              >
                {{ $t('base.common.error_retry') }}
              </div>
            </slot>

            <slot v-else-if="!results?.length">
              <RkComboboxEmpty>
                <div class="dark:text-gray-400 px-2xl py-md font-semibold">
                  {{ $t('base.common.no_results') }}
                </div>
              </RkComboboxEmpty>
            </slot>

            <template v-else-if="results?.length && groupBy">
              <RkComboboxGroup
                v-for="[groupKey, groupItems] in groupedResults"
                :key="groupKey as string"
              >
                <RkComboboxLabel class="px-2xl py-md  dark:text-gray-400">
                  <slot
                    name="result-group-label"
                    :label="groupKey"
                  />
                </RkComboboxLabel>

                <RkComboboxItem
                  v-for="result in groupItems"
                  :key="JSON.stringify(result)"
                  as-child
                  :value="JSON.stringify(result)"
                  class="dark:data-[highlighted]:bg-gray-900 px-2xl py-md font-semibold"
                  @select.prevent
                >
                  <slot
                    name="result-item"
                    :result
                  />
                </RkComboboxItem>
              </RkComboboxGroup>
            </template>

            <template v-else-if="results?.length && !groupBy">
              <RkComboboxItem
                v-for="result in results"
                :key="JSON.stringify(result)"
                as-child
                :value="JSON.stringify(result)"
                class="dark:data-[highlighted]:bg-gray-900 px-2xl py-md font-semibold"
                @select.prevent
              >
                <slot
                  name="result-item"
                  :result
                />
              </RkComboboxItem>
            </template>
          </div>
        </RkComboboxViewport>
      </RkComboboxContent>
    </RkComboboxRoot>
  </form>
</template>

<style lang="scss" scoped>
.search-input::-webkit-search-cancel-button {
  display: none;
}

form {
  position: relative;

  &:before {
    position: absolute;
    content: '';
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border-radius: var(--radius-3xl);
    background: var(--color-black);
    opacity: 0.2;
  }
}
</style>
