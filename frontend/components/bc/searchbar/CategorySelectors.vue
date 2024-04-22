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
  liveState.value.set(category, selected)
  emit('change')
}
</script>

<template>
  <div>
    <span v-for="filter of liveState" :key="filter[0]">
      <label class="filter-button">
        <input
          type="checkbox"
          class="hidden-checkbox"
          :true-value="true"
          :false-value="false"
          :checked="filter[1]"
          :onchange="(e:any) => selectionHasChanged(filter[0], e.target.checked)"
        >
        <span class="face" :class="barStyle">
          {{ t(...CategoryInfo[filter[0]].filterLabel) }}
        </span>
      </label>
    </span>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.filter-button {
  @include fonts.small_text_bold;

  .face{
    display: inline-block;
    cursor: pointer;
    border-radius: 10px;
    height: 17px;
    padding-top: 2.5px;
    padding-left: 8px;
    padding-right: 8px;
    text-align: center;
    margin-right: 6px;
    transition: 0.2s;
    margin-bottom: 8px;

    &.gaudy {
      color: var(--primary-contrast-color);
      background-color: var(--searchbar-filter-unselected-gaudy);
    }
    &.discreet {
      color: var(--light-black);
      background-color: var(--light-grey);
    }

    &:hover {
      background-color: var(--light-grey-3);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }

  .hidden-checkbox {
    display: none;
    width: 0;
    height: 0;
    &:checked + .face {
      background-color: var(--button-color-active);
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }
}
</style>
