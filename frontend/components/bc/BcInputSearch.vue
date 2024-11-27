<script setup lang="ts" generic="T">
const props = withDefaults(defineProps<{
  hasError?: boolean,
  hasFocus?: boolean,
  isLoading: boolean,
  results: T[] | undefined,
  shouldClearOnSubmit?: boolean,
  shouldSelectFirstResult?: boolean,
}>(), {
  shouldClearOnSubmit: true,
})

const emit = defineEmits<{
  (e: 'search', input: string): void,
  (e: 'submit', value: T): void,
}>()

const { t: $t } = useTranslation()

const input = defineModel<string>()

const elementInput = ref<HTMLElement >()
const hasInput = computed(() => input.value?.length)

const idResults = useId()
const elementResults = ref<HTMLElement>()

const idElementList = useId()
const elementList = ref<Array<HTMLElement>>()

const hasResults = computed(() => props.results?.length)
const hasPopover = ref(false)

onMounted(() => {
  if (props.hasFocus) elementInput.value?.focus()
})

const currentIndex = ref(-1)

const increaseCurrentIndex = () => {
  if (!props.results?.length) return
  showPopover()
  currentIndex.value = currentIndex.value + 1
  if (currentIndex.value >= props.results.length) {
    currentIndex.value = 0
  }
}

const decreaseCurrentIndex = () => {
  if (!props.results?.length) return
  showPopover()
  currentIndex.value = currentIndex.value - 1
  if (currentIndex.value < 0) {
    currentIndex.value = props.results.length - 1
  }
}

const {
  bottom,
  left,
  right,
  update: updatePositionOfInput,
} = useElementBounding(elementInput, {
  immediate: false,
})

const showPopover = () => {
  updatePositionOfInput()
  hasPopover.value = true
  if (!elementResults.value || elementResults.value.matches(':popover-open')) return
  if (props.shouldSelectFirstResult) currentIndex.value = 0
  elementResults.value.showPopover()
}
const hidePopover = () => {
  hasPopover.value = false
  currentIndex.value = -1
  if (!elementResults.value || !elementResults.value.matches(':popover-open')) return
  elementResults.value.hidePopover()
}

const isEmpty = computed(() => !props.results?.length)
watchDebounced(input, async () => {
  emit('search', input.value ?? '')
})
watch(hasInput, () => {
  if (!hasInput.value) {
    hidePopover()
  }
})
const hasErrorOrNoResults = computed(() => props.hasError || props.results === undefined)
watch([
  () => props.results,
  () => props.hasError,
], () => {
  if (!hasInput.value) return
  showPopover()
})
watchEffect(() => {
  elementResults.value?.style.setProperty('--position-top', bottom.value + 'px')
  elementResults.value?.style.setProperty('--position-left', left.value + 'px')
  elementResults.value?.style.setProperty('--position-right', right.value + 'px')
})

const handleEsc = (event: KeyboardEvent) => {
  if (!hasInput.value) return
  event.stopPropagation()
  if (!hasPopover.value) return
  event.preventDefault()
  hidePopover()
}
const handleFocus = (event: Event) => {
  if (!(event.target as HTMLInputElement).value) return
  if (hasResults.value) showPopover()
}

const selectedResult = computed(() => props.results?.[currentIndex.value])
const handleSubmit = () => {
  if (!selectedResult.value) return
  emit('submit', toRaw(selectedResult.value))
  if (props.shouldClearOnSubmit) input.value = ''
  hidePopover()
}
const handleClick = (index: number) => {
  currentIndex.value = index
  elementInput.value?.focus()
  handleSubmit()
}
const isResultsHovered = () => elementResults.value?.matches(':hover')
const handleBlur = () => {
  if (isResultsHovered()) {
    return
  }
  hidePopover()
}

const ariaActivedescendant = computed(() => {
  if (props.hasError) return `${idResults}-error`
  if (isEmpty.value) return `${idResults}-empty`
  return `${idElementList}-${currentIndex.value}`
})
</script>

<template>
  <form
    role="search"
    class="bc-input-search__form"
    @keydown.arrow-up.prevent="hasResults && decreaseCurrentIndex()"
    @keydown.arrow-down.prevent="hasResults && increaseCurrentIndex()"
    @keydown.arrow-down.alt.exact.prevent="hasResults && showPopover()"
    @keydown.esc="handleEsc"
    @submit.prevent="handleSubmit"
  >
    <input
      ref="elementInput"
      v-model.trim="input"
      role="combobox"
      :aria-expanded="hasPopover"
      :aria-busy="isLoading"
      aria-autocomplete="none"
      :aria-controls="idResults"
      :aria-label="$t('dashboard.validator.management.search.label')"
      :aria-activedescendant
      :placeholder="$t('dashboard.validator.management.search.placeholder')"
      type="search"
      class="bc-input-search__input"
      :class="{
        'bc-input-search__input--has-popover': hasPopover,
        'bc-input-search__input--is-loading': isLoading,
      }"
      @blur="handleBlur"
      @focus="handleFocus"
    >
    <span
      class="bc-input-search__loading-indicator"
    >
      <BcLoadingSpinner
        v-if="isLoading"
        size="full"
        loading
      />
    </span>
    <div
      :id="idResults"
      ref="elementResults"
      popover="manual"
      class="bc-input-search__results"
      :class="{ 'bc-input-search__results--has-popover': hasPopover }"
    >
      <div
        class="bc-input-search__results_content"
      >
        <div
          v-if="hasErrorOrNoResults"
          :id="`${idResults}-error`"
          class="bc-input-search__list-content-error"
        >
          <slot name="error">
            {{ $t('common.search_error') }}
          </slot>
        </div>
        <div
          v-else-if="isEmpty"
          :id="`${idResults}-empty`"
          class="bc-input-search__list-content-empty"
        >
          <slot name="empty">
            {{ $t('common.search_empty') }}
          </slot>
        </div>
        <slot
          v-else
          name="results"
          :results
          role="listbox"
        >
          <ul
            role="listbox"
            :aria-label="$t('dashboard.validator.management.search.add_validators')"
            class="bc-input-search__list"
          >
            <BcInputSearchItem
              v-for="(item, index) in results"
              :id="`${idElementList}-${index}`"
              :key="`${item}-${index}`"
              ref="elementList"
              :is-aria-selected="currentIndex === index"
              tabindex="-1"
              @click="handleClick(index)"
            >
              <slot name="result" :item />
            </BcInputSearchItem>
          </ul>
        </slot>
      </div>
    </div>
  </form>
</template>

<style scoped lang="scss">
.bc-input-search__form {
  --padding: .2813rem;
  position: relative;
}
.bc-input-search__input {
  width: 100%;
  flex: 1;
  padding: var(--padding);
  border: none;
  border: .0625rem solid var(--input-border-color);
  border-radius: var(--corner-radius, .25rem);
  background-color: var(--input-background);
  color: var(--input-active-text-color);
  transition: border-radius 200ms ease-in;

  &::-webkit-search-cancel-button {
    display: none;
  }

  &:focus-visible,
  &:focus-within,
  &--has-popover
  {
    border-color: var(--primary-color);
    outline: none;
  }
}
.bc-input-search__input--is-loading {
  padding-right: 1.625rem;
}
.bc-input-search__input--has-popover {
    --missing-border-bottom: .0625rem;
    border-bottom: unset;
    padding-bottom: calc(var(--padding) + var(--missing-border-bottom) );
    border-bottom-left-radius: unset;
    border-bottom-right-radius: unset;
}
.bc-input-search__loading-indicator{
  position: absolute;
  inset: var(--padding);
  left: unset;
  aspect-ratio: 1;
}
.bc-input-search__results {
  --position-top: 0;
  --position-left: 0;
  --position-right: 0;
  border: .0625rem solid var(--primary-color);
  border-top: none;
  border-bottom-left-radius: var(--corner-radius, .25rem);
  border-bottom-right-radius: var(--corner-radius, .25rem);
  background-color: var(--input-background);
  inset: unset;
  top: calc(var(--position-top) - .0625rem);
  left: var(--position-left);
  width: calc(var(--position-right) - var(--position-left));
  color: var(--input-active-text-color);
  padding: 0;
  overflow: clip;
  height: 0;
  transition: height 200ms ease-in, display 200ms;
  transition-delay: 200ms;
  height: auto;
  &:popover-open{
    @starting-style {
      height: 0;
    }
  }
}
.bc-input-search__results_content {
  border-top: .0625rem solid var(--input-border-color);
}
.bc-input-search__list {
  overflow: auto;
  max-height: 13.75rem;
  list-style: none;
  padding-inline-start: 0;
  display: grid;
  grid-auto-flow: row;
  grid-template-columns: min-content max-content 2fr max-content;
  column-gap: .625rem;
}
.bc-input-search__list-content-error,
.bc-input-search__list-content-empty {
  padding: var(--padding);
  text-align: center;
  color: var(--text-color-disabled);
}
</style>
