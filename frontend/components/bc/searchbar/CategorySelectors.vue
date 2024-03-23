<script setup lang="ts">
import {
  Category,
  CategoryInfo,
  TypeInfo,
  getListOfResultTypesInCategory,
  type SearchBarStyle
} from '~/types/searchbar'

const { t: $t } = useI18n()
const emit = defineEmits(['change'])
const props = defineProps<{
    initialState: Record<string, boolean>, // each key is a category name (as enumerated in Category in searchbar.ts)
    barStyle: SearchBarStyle
 }>()

let componentIsReady = false
const state = ref<Record<string, boolean>>({}) // each key is a category name (as enumerated in Category in searchbar.ts)

onMounted(() => {
  componentIsReady = false
  state.value = { ...props.initialState }
  componentIsReady = true
})

function selectionHasChanged () {
  if (componentIsReady) { // ensures that we do not emit change-events during the initialization of the buttons (see above)
    emit('change', state.value)
  }
}

function tellWhatFilterDoes (category : Category) : string {
  let hint = $t('search_bar.shows') + ' '

  if (category === Category.Validators) {
    hint += $t('search_bar.this_type') + ' '
    hint += 'Validator'
  } else {
    const list = getListOfResultTypesInCategory(category, false)

    hint += (list.length === 1 ? $t('search_bar.this_type') : $t('search_bar.these_types')) + ' '
    for (let i = 0; i < list.length; i++) {
      hint += TypeInfo[list[i]].title
      if (i < list.length - 1) {
        hint += ', '
      }
    }
  }

  return hint
}
</script>

<template>
  <div>
    <span v-for="filter of Object.keys(state)" :key="filter">
      <BcTooltip :text="tellWhatFilterDoes(filter as Category)">
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
            {{ CategoryInfo[filter as Category].filterLabel }}
          </span>
        </label>
      </BcTooltip>
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
    margin-left: 6px;
    transition: 0.2s;

    &.discreet {
      color: var(--light-black);
      background-color: var(--light-grey);
    }
    &.gaudy {
      color: var(--primary-contrast-color);
      background-color: var(--searchbar-filter-unselected-gaudy);
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
