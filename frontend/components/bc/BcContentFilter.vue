<script lang="ts" setup>
import { faMagnifyingGlass } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const props = defineProps<{
  disabledFilter?: boolean,
  isLoading?: boolean,
  searchPlaceholder?: string,
}>()

const emit = defineEmits<{ (e: 'filter-changed', value: string): void }>()

const isFilterVisible = ref(false)
const filter = ref('')

const button = ref<null | { $el: HTMLButtonElement }>(null)

const input = ref<null | { $el: HTMLInputElement }>(null)
const focusAndSelect = (inputElement: { $el: HTMLInputElement }) => {
  if (inputElement?.$el) {
    // make sure the input is not disabled anymore
    nextTick(() => {
      inputElement.$el.focus()
      inputElement.$el.select()
    })
  }
}
const closeFilter = () => {
  isFilterVisible.value = false
  button.value?.$el.focus()
}
const handleClick = () => {
  isFilterVisible.value = !isFilterVisible.value
  if (input.value) {
    focusAndSelect(input.value)
  }
}
watchDebounced(filter, () => {
  emit('filter-changed', filter.value)
})
</script>

<template>
  <div class="filter_elements_container">
    <InputText
      ref="input"
      v-model.trim="filter"
      type="search"
      aria-busy="true"
      :placeholder="props.searchPlaceholder"
      :disabled="!isFilterVisible"
      :class="{ visible: isFilterVisible }"
      @keydown.escape.stop="closeFilter"
    />
    <Button
      ref="button"
      :disabled="disabledFilter"
      :aria-expanded="isFilterVisible"
      severity="secondary"
      class="p-button-icon-only bc-content-filter__button"
      :class="{ filter_visible: isFilterVisible }"
      @click="handleClick"
    >
      <BcScreenreaderOnly>
        {{ isFilterVisible ? $t('filter.open') : $t('filter.close') }}
      </BcScreenreaderOnly>
      <BcLoadingSpinner
        v-if="isFilterVisible && isLoading"
        size="full"
        alignment="center"
        loading
      />
      <FontAwesomeIcon
        v-else
        :icon="faMagnifyingGlass"
      />
    </Button>
  </div>
</template>

<style lang="scss">
.filter_elements_container {
  --outline-width: 0.125rem;
  --outline-offset: 0.125rem;
  padding: .25rem;
  display: flex;
  justify-content: flex-end;
  position: relative;

  > .p-inputtext:first-child {
    height: var(--default-button-height);
    width: 0;
    opacity: 0;
    padding: 0;
    position: absolute;
    transition: width 0.2s ease-in-out, opacity 0.01s ease-in-out 0.19s,
    padding 0.2s ease-in-out;
    &.visible {
      width: 230px;
      opacity: 100%;
      padding: 4px;

      transition: width 0.2s ease-in-out, opacity 0.01s ease-in-out,
        padding 0.2s ease-in-out;
    }
  }

  > :last-child {
    flex-shrink: 0;
    border-top-left-radius: var(--border-radius);
    border-bottom-left-radius: var(--border-radius);
    transition: all 0.2s ease-in-out;

    &.filter_visible {
      border-top-left-radius: 0;
      border-bottom-left-radius: 0;
    }
  }
}
button.bc-content-filter__button:focus-visible {
  outline: var(--outline-width) solid var(--blue-500);
  outline-offset: var(--outline-offset);
}
</style>
