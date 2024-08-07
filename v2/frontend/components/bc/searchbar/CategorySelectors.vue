<script setup lang="ts">
import {
  type Category,
  type SearchbarShape,
  CategoryInfo,
  type SearchbarColors,
  type SearchbarDropdownLayout,
  type CategoryFilter,
} from '~/types/searchbar'

const emit = defineEmits<{ (e: 'change'): void }>()
defineProps<{
  barShape: SearchbarShape
  colorTheme: SearchbarColors
  dropdownLayout: SearchbarDropdownLayout
}>()
const liveState = defineModel<CategoryFilter>({ required: true }) // each entry has a Category as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

const { t } = useTranslation()

function selectionHasChanged(category: Category, selected: boolean) {
  liveState.value.set(category, selected)
  emit('change')
}
</script>

<template>
  <div class="group">
    <BcSearchbarFilterButton
      v-for="filter of liveState"
      :key="filter[0]"
      :state="filter[1]"
      class="button"
      :bar-shape="barShape"
      :color-theme="colorTheme"
      :dropdown-layout="dropdownLayout"
      @change="(selected : boolean) => selectionHasChanged(filter[0], selected)"
    >
      {{ t(...CategoryInfo[filter[0]].filterLabel) }}
    </BcSearchbarFilterButton>
  </div>
</template>

<style lang="scss" scoped>
.group {
  display: inline-block;
  vertical-align: top;
  .button {
    margin-right: var(--padding-small);
    margin-bottom: 8px;
  }
}
</style>
