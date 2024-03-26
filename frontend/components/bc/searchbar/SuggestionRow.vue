<script setup lang="ts">
import { ChainIDs } from '~/types/networks'
import {
  CategoryInfo,
  SubCategoryInfo,
  TypeInfo,
  ResultType,
  isOutputAnAPIresponse,
  type ResultSuggestion,
  type SearchBarStyle,
  SearchBarPurpose
} from '~/types/searchbar'

const emit = defineEmits(['row-selected'])
const props = defineProps<{
    suggestion: ResultSuggestion,
    chainId: ChainIDs,
    resultType: ResultType,
    barStyle: SearchBarStyle,
    barPurpose: SearchBarPurpose
}>()

function formatEmbeddedNameCell () : string {
  let label : string

  if (props.barPurpose === SearchBarPurpose.Accounts) {
    label = SubCategoryInfo[TypeInfo[props.resultType].subCategory].title
  } else {
    label = props.suggestion.output.name
  }

  if (props.suggestion.count >= 2) {
    return String(props.suggestion.count) + ' ' + label + 's'
  }

  return label
}

function formatEmbeddedDescriptionCell () : string {
  if (isOutputAnAPIresponse(props.resultType, 'description')) {
    // we tell the user what is the data that they see (ex: "Index" for a validator index)
    switch (props.resultType) {
      case ResultType.ValidatorsByIndex :
      case ResultType.ValidatorsByPubkey :
        return 'Index ' + props.suggestion.output.description
      // more cases might arise in the future
    }
  }
  return props.suggestion.output.description
}

function formatEmbeddedLowleveldataCell () : string {
  if (props.suggestion.nameWasUnknown || !isOutputAnAPIresponse(props.resultType, 'name')) {
    return props.suggestion.output.lowLevelData
  }
  return props.suggestion.output.name
}
</script>

<template>
  <div
    v-if="barStyle == 'embedded'"
    class="row-common row-embedded"
    :class="barStyle"
    @click="emit('row-selected', chainId, resultType, suggestion.queryParam, suggestion.count)"
  >
    <div v-if="chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon not-alone" />
      <IconNetwork :chain-id="chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon alone" />
    </div>
    <div class="cell-name" :class="barStyle">
      {{ formatEmbeddedNameCell() }}
    </div>
    <div class="cell-blockchaininfo-common cell-bi-lowleveldata" :class="barStyle">
      <BcSearchbarMiddleEllipsis :width-is-fixed="true">
        {{ formatEmbeddedLowleveldataCell() }}
      </BcSearchbarMiddleEllipsis>
    </div>
    <div v-if="suggestion.output.description !== ''" class="cell-blockchaininfo-common cell-bi-description" :class="barStyle">
      {{ formatEmbeddedDescriptionCell() }}
    </div>
  </div>

  <div
    v-else
    class="row-common row-gaudyordiscreet"
    :class="barStyle"
    @click="emit('row-selected', chainId, resultType, suggestion.queryParam, suggestion.count)"
  >
    <div v-if="chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon not-alone" />
      <IconNetwork :chain-id="chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon alone" />
    </div>
    <div class="cell-name" :class="barStyle">
      <BcSearchbarMiddleEllipsis>{{ suggestion.output.name }}</BcSearchbarMiddleEllipsis>
    </div>
    <div class="cell-blockchaininfo" :class="barStyle">
      <span v-if="suggestion.output.description !== ''" class="cell-bi-description" :class="barStyle">
        <BcSearchbarMiddleEllipsis>{{ suggestion.output.description }}</BcSearchbarMiddleEllipsis>
      </span>
      <span v-if="suggestion.output.lowLevelData !== ''" class="cell-bi-lowleveldata" :class="[barStyle,(suggestion.output.description !== '')?'greyish':'']">
        <BcSearchbarMiddleEllipsis :width-is-fixed="true">{{ suggestion.output.lowLevelData }}</BcSearchbarMiddleEllipsis>
      </span>
    </div>
    <div class="cell-category" :class="barStyle">
      <span class="category-label" :class="barStyle">
        {{ CategoryInfo[TypeInfo[resultType].category].title }}
      </span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

// styles common to all modes

.row-common {
  cursor: pointer;
  display: grid;
  min-width: 0;
  margin-left: 4px;
  padding-left: 4px;
  margin-right: 4px;
  padding-right: 4px;
  padding-top: 7px;
  padding-bottom: 7px;
  border-radius: var(--border-radius);

  &:hover {
    &.discreet {
      background-color: var(--searchbar-background-hover-discreet);
    }
    &.gaudy,
    &.embedded {
      background-color: var(--dropdown-background-hover);
    }
  }
  &:active {
    &.discreet {
      background-color: var(--searchbar-background-pressed-discreet);
    }
    &.gaudy,
    &.embedded {
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

  .cell-name {
    grid-column: 2;
    grid-row: 1;
    display: inline-block;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }
}

// specific style for the embedded mode

.row-embedded {
  @media (min-width: 600px) { // large screen
    grid-template-columns: 40px 106px auto min-content;
  }
  @media (max-width: 600px) { // mobile
    grid-template-columns: 40px auto min-content;
  }

  .cell-name {
    @media (min-width: 600px) { // large screen
      font-weight: var(--roboto-medium);
    }
    @media (max-width: 600px) { // mobile
      font-weight: var(--roboto-regular);
    }
  }

  .cell-blockchaininfo-common {
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
    white-space: nowrap;
    grid-row: 1;
  }

  .cell-bi-description {
    position: relative;
    @media (min-width: 600px) { // large screen
      grid-column: 4;
    }
    @media (max-width: 600px) { // mobile
      grid-column: 3;
      color: var(--searchbar-text-detail-gaudy);
    }
    width: 100px;
    margin-left: auto;
    justify-content: right;
  }

  .cell-bi-lowleveldata {
    position: relative;
    @media (min-width: 600px) { // large screen
      grid-row: 1;
      grid-column: 3;
      font-weight: var(--roboto-medium);
    }
    @media (max-width: 600px) { // mobile
      grid-row: 2;
      grid-column-end: span 2;
      font-weight: var(--roboto-regular);
    }
    width: 100%;
  }
}

// specific style for the gaudy and discreet modes

.row-gaudyordiscreet {
  @media (min-width: 600px) { // large screen
    &.gaudy {
      grid-template-columns: 40px 106px auto 114px;
    }
    &.discreet {
      grid-template-columns: 40px 114px auto;
    }
  }
  @media (max-width: 600px) { // mobile
    grid-template-columns: 40px 114px auto;
  }

  .cell-name {
    font-weight: var(--roboto-medium);
  }

  .cell-blockchaininfo {
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
    font-weight: var(--roboto-medium);
    white-space: nowrap;

    .cell-bi-description {
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

    .cell-bi-lowleveldata {
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
    position: relative;
    @media (min-width: 600px) { // large screen
      &.gaudy {
        grid-column: 4;
        grid-row: 1;
        margin-top: auto;
        margin-bottom: auto;
        margin-left: auto;
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
      display: inline-block;
      position: relative;
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
