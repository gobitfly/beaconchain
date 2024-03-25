<script setup lang="ts">
import { ChainIDs } from '~/types/networks'
import {
  CategoryInfo,
  TypeInfo,
  ResultType,
  type ResultSuggestion,
  type SearchBarStyle
} from '~/types/searchbar'

const emit = defineEmits(['row-selected'])
defineProps<{
    suggestion: ResultSuggestion,
    chainId: ChainIDs,
    resultType: ResultType,
    barStyle: SearchBarStyle
}>()
</script>

<template>
  <div class="row" :class="barStyle" @click="emit('row-selected', chainId, resultType, suggestion.queryParam, suggestion.count)">
    <span v-if="chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon not-alone" />
      <IconNetwork :chain-id="chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </span>
    <span v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon alone" />
    </span>
    <span class="cell-0" :class="barStyle">
      <BcSearchbarMiddleEllipsis>{{ suggestion.output[0] }}</BcSearchbarMiddleEllipsis>
    </span>
    <span class="cell-1and2" :class="barStyle">
      <span v-if="suggestion.output[1] !== ''" class="cell-1" :class="barStyle">
        <BcSearchbarMiddleEllipsis>{{ suggestion.output[1] }}</BcSearchbarMiddleEllipsis>
      </span>
      <span v-if="suggestion.output[2] !== ''" class="cell-2" :class="[barStyle,(suggestion.output[1] !== '')?'greyish':'']">
        <BcSearchbarMiddleEllipsis v-if="TypeInfo[resultType].dropdownOutput[1] === undefined" :width-is-fixed="true">({{ suggestion.output[2] }})</BcSearchbarMiddleEllipsis>
        <BcSearchbarMiddleEllipsis v-else :width-is-fixed="true">{{ suggestion.output[2] }}</BcSearchbarMiddleEllipsis>
      </span>
    </span>
    <span class="cell-category" :class="barStyle">
      <span class="category-label" :class="barStyle">
        {{ CategoryInfo[TypeInfo[resultType].category].filterLabel }}
      </span>
    </span>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.row {
  cursor: pointer;
  display: grid;
  min-width: 0;
  right: 0px;
  padding-top: 7px;
  padding-bottom: 7px;
  @media (min-width: 600px) { // large screen
    &.gaudy {
      grid-template-columns: 40px 106px 488px auto;
      padding-left: 4px;
      padding-right: 4px;
    }
    &.discreet {
      grid-template-columns: 40px 106px 298px;
    }
  }
  @media (max-width: 600px) { // mobile
    grid-template-columns: 40px 106px 218px;
  }
  border-radius: var(--border-radius);

  &:hover {
    &.discreet {
      background-color: var(--searchbar-background-hover-discreet);
    }
    &.gaudy {
      background-color: var(--dropdown-background-hover);
    }
  }
  &:active {
    &.discreet {
      background-color: var(--searchbar-background-pressed-discreet);
    }
    &.gaudy {
      background-color: var(--button-color-pressed);
    }
  }

  .cell-icons {
    position: relative;
    grid-column: 1;
    grid-row: 1;
    @media (max-width: 600px) { // mobile
      grid-row-end: span 2;
    }
    &.discreet {
      grid-row-end: span 2;
    }
    display: flex;
    margin-top: auto;
    margin-bottom: auto;
    width: 30px;
    height: 36px;

    .type-icon {
      &.not-alone {
        display: inline;
        position: relative;
        top: 2px;
      }
      &.alone {
        display: flex;
        margin-top: auto;
        margin-bottom: auto;
      }
      width: 20px;
      max-height: 20px;
    }
    .network-icon {
      position: absolute;
      bottom: 0px;
      right: 0px;
      width: 20px;
      height: 20px;
    }
  }

  .cell-0 {
    grid-column: 2;
    grid-row: 1;
    display: inline-block;
    position: relative;
    margin-top: auto;
    &.gaudy {
      margin-bottom: auto;
    }
    margin-right: 14px;
    left: 0px;
    font-weight: var(--roboto-medium);
  }

  .cell-1and2 {
    grid-column: 3;
    grid-row: 1;
    display: flex;
    @media (max-width: 600px) { // mobile
      grid-row-end: span 2;
      flex-direction: column;
    }
    &.discreet {
      grid-row-end: span 2;
      flex-direction: column;
    }
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
    left: 0px;
    font-weight: var(--roboto-medium);
    white-space: nowrap;

    .cell-1 {
      display: flex;
      max-width: 100%;
      @media (min-width: 600px) { // large screen
        &.gaudy {
          max-width: 27%;
        }
      }
      position: relative;
      margin-right: 0.8ch;
    }

    .cell-2 {
      display: flex;
      position: relative;
      flex-grow: 1;
      &.greyish.discreet {
        color: var(--searchbar-text-detail-discreet);
      }
      &.greyish.gaudy {
        color: var(--searchbar-text-detail-gaudy);
      }
    }
  }

  .cell-category {
    display: block;
    @media (min-width: 600px) { // large screen
      &.gaudy {
        grid-column: 4;
        grid-row: 1;
        margin-top: auto;
        margin-bottom: auto;
        margin-right: 2px;
        float: right;
        justify-content: right;
        text-align: right;
      }
      &.discreet {
        grid-column: 2;
        grid-row: 2;
      }
    }
    @media (max-width: 600px) { // mobile
      grid-column: 2;
      grid-row: 2;
    }
    .category-label {
      &.discreet {
        color: var(--searchbar-text-detail-discreet);
      }
      &.gaudy {
        color: var(--searchbar-text-detail-gaudy);
      }
    }
  }
}
</style>
