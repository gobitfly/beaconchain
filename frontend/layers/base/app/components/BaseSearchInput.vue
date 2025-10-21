<script setup lang="ts" generic="SearchResultType extends Record<string, any>">
import type { PointerDownOutsideEvent } from 'reka-ui'

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
    emit('search', input.value)
  },
  {
    immediate: false,
  },
)

const groupedResults = computed(() => {
  if (!results?.length) return
  if (!groupBy) return

  const groupedResults = Object.groupBy(results, result => result[groupBy] as PropertyKey)

  return Object.entries(groupedResults) as [SearchResultType[GroupBy], SearchResultType[]][]
})

const searchInput = useTemplateRef('search-input')
const handleClickOutside = (e: PointerDownOutsideEvent) => {
  if (e.target === searchInput.value?.$el) return

  input.value = ''
  hasSearched.value = false
}
const idSearchInput = useId()
</script>

<template>
  <form
    role="search"
    class="base-search-input__form p-2xl isolate"
  >
    <RkComboboxRoot
      :open-on-focus="!!results?.length"
      class="relative"
      ignore-filter
      :reset-search-term-on-blur="false"
    >
      <RkLabel
        :for="idSearchInput"
        class="absolute bottom-2xl left-2xl dark:text-gray-400 text-sm-tight"
      >
        {{ label }}
      </RkLabel>
      <RkComboboxInput
        :id="idSearchInput"
        ref="search-input"
        v-model.trim="input"
        type="search"
        :aria-busy="isLoading"
        :placeholder
        class="w-full text-2xl font-semibold rounded-3xl pt-3xl pr-5xl pb-6xl pl-2xl dark:bg-gray-950
        bg-white dark:focus:bg-black placeholder:dark:text-gray-500 dark:text-white border border-gray-500 dark:border-charcoal-500
        dark:focus:border-charcoal-50 dark:focus-within:outline-0"
        @update:model-value="(value) => { if (!value) hasSearched = false }"
      />
      <RkComboboxContent
        v-if="results !== undefined || isLoading || hasError"
        class="absolute z-10 bg-gray-50 dark:bg-gray-950 mt-xl rounded-xl w-full max-h-[400px]"
        @pointer-down-outside="handleClickOutside"
        @focus-outside.prevent
      >
        <slot
          name="dropdown-fixed-header"
          :id-search-input
        />
        <div
          role="presentation"
          class="overflow-y-auto overscroll-contain"
          tabindex="-1"
        >
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
                class="px-2xl py-md dark:text-gray-400 flex items-center"
              >
                <div>
                  {{ $t('base.common.something_went_wrong') }}
                </div>
                <BaseButton
                  trailing-icon="rotate"
                  variant="quaternary"
                  @click="$emit('search', input)"
                >
                  {{ $t('base.common.action.try_again') }}
                </BaseButton>
              </div>
            </slot>

            <div
              v-else-if="!results?.length"
              class="dark:text-gray-400 px-2xl py-md font-semibold"
            >
              {{ $t('base.common.no_results') }}
            </div>

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
        </div>
      </RkComboboxContent>
    </RkComboboxRoot>
  </form>
</template>

<style lang="scss" scoped>
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
