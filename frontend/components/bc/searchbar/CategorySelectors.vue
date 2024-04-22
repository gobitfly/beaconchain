<script setup lang="ts">
import {
  Category,
  CategoryInfo,
  SearchbarStyle,
  type CategoryFilter
} from '~/types/searchbar'

const emit = defineEmits<{(e: 'change') : void}>()
defineProps<{
  barStyle: SearchbarStyle
}>()
const liveState = defineModel<CategoryFilter>({ required: true }) // each entry has a Category as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

const { t } = useI18n()

function selectionHasChanged (category : Category, selected : boolean) {
  liveState.value.set(category, selected) // the map element cannot save the infornmation from `v-model`, so we do it with .set()
  emit('change')
}
</script>

<template>
  <div>
    <BcSearchbarMiniButton
      v-for="filter of liveState"
      :key="filter[0]"
      v-model="filter[1]"
      class="button"
      :bar-style="barStyle"
      @change="(activ : boolean) => selectionHasChanged(filter[0], activ)"
    >
      {{ t(...CategoryInfo[filter[0]].filterLabel) }}
    </BcSearchbarMiniButton>
  </div>
</template>

<style lang="scss" scoped>
.button {
  margin-right: 6px;
  margin-bottom: 8px;
}
</style>
