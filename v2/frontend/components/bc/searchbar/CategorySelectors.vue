<script setup lang="ts">
import {
  Category,
  CategoryInfo,
  SearchbarStyle
} from '~/types/searchbar'

const emit = defineEmits(['change'])
const props = defineProps<{
    initialState: Record<string, boolean>, // each field has a stringifyEnum(Category) as key and the state of the option as value
    barStyle: SearchbarStyle
 }>()

let componentIsReady = false
const state = ref<Record<string, boolean>>({}) // each field has a stringifyEnum(Category) as key and the state of the option as value

onMounted(() => {
  componentIsReady = false
  state.value = { ...props.initialState }
  componentIsReady = true
})

function selectionHasChanged () {
  if (componentIsReady) { // ensures that we do not emit change-events during the initialization of the buttons (see the code in onMounted)
    console.log('Category selector')
    emit('change', state.value)
  }
}
</script>

<template>
  <div>
    <span v-for="filter of Object.keys(state)" :key="filter">
      <label class="filter-button">
        <input
          v-model="state[filter]"
          type="checkbox"
          class="hiddencheckbox"
          :true-value="true"
          :false-value="false"
          @change="selectionHasChanged"
        >
        <span class="face" :class="barStyle">
          {{ CategoryInfo[Number(filter) as Category].filterLabel }}
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
    border-radius: 10px;
    height: 17px;
    padding-top: 2.5px;
    padding-left: 8px;
    padding-right: 8px;
    text-align: center;
    margin-right: 6px;
    transition: 0.2s;

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

  .hiddencheckbox {
    display: none;
    width: 0;
    height: 0;
  }

  .hiddencheckbox:checked + .face {
    background-color: var(--button-color-active);
    &:hover {
      background-color: var(--button-color-hover);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }
}
</style>
