<script setup lang="ts">
import {
  Category,
  CategoryInfo,
  SearchbarStyle
} from '~/types/searchbar'

const emit = defineEmits(['change'])
const props = defineProps<{
  liveState: Record<string, boolean>, // each field has a stringifyEnum(Category) as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.
  barStyle: SearchbarStyle
}>()

const { t } = useI18n()

const stateRef = ref<Record<string, boolean>>({}) // each field has a stringifyEnum(Category) as key and the state of the option as value

onMounted(() => {
  stateRef.value = props.liveState
})
watch(props, () => {
  stateRef.value = props.liveState
})

function selectionHasChanged () {
  emit('change')
}
</script>

<template>
  <div>
    <span v-for="filter of Object.keys(stateRef)" :key="filter">
      <label class="filter-button">
        <input
          v-model="stateRef[filter]"
          type="checkbox"
          class="hiddencheckbox"
          :true-value="true"
          :false-value="false"
          @change="selectionHasChanged"
        >
        <span class="face" :class="barStyle">
          {{ t(...CategoryInfo[Number(filter) as Category].filterLabel) }}
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
