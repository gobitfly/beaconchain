<script lang="ts" setup>
import { faMagnifyingGlass } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type InputText from 'primevue/inputtext'

interface Props {
  searchPlaceholder?: string;
  disabledFilter?: boolean;
}
const props = defineProps<Props>()

defineEmits<{(e: 'filter-changed', value: string): void }>()

const isFilterVisible = ref(false)
const filter = ref('')

const button = ref<{$el: HTMLButtonElement} | null>(null)

const input = ref<{$el: HTMLInputElement} | null>(null)
const focusAndSelect = (inputElement: {$el: HTMLInputElement}) => {
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
</script>

<template>
  <div class="filter_elements_container">
    <InputText
      ref="input"
      v-model="filter"
      :placeholder="props.searchPlaceholder"
      :disabled="!isFilterVisible"
      :class="{ visible: isFilterVisible }"
      @keydown.escape.stop="closeFilter"
      @input="$emit('filter-changed', filter)"
    />
    <Button
      ref="button"
      :disabled="disabledFilter"
      :aria-expanded="isFilterVisible"
      data-secondary
      class="p-button-icon-only"
      :class="{ filter_visible: isFilterVisible }"
      @click="handleClick"
    >
      <span class="sr-only">
        {{ !isFilterVisible ? $t('filter.open') : $t('filter.close') }}
      </span>
      <FontAwesomeIcon :icon="faMagnifyingGlass" />
    </Button>
  </div>
</template>

<style lang="scss">
.filter_elements_container {
  display: flex;
  justify-content: flex-end;
  position: relative;

  > :first-child {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
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
</style>
