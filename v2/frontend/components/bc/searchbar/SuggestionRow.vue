<script setup lang="ts">
import { ChainIDs } from '~/types/networks'
import {
  CategoryInfo,
  SubCategoryInfo,
  TypeInfo,
  ResultType,
  isOutputAnAPIresponse,
  type ResultSuggestion,
  SearchbarStyle,
  SearchbarPurpose
} from '~/types/searchbar'

const props = defineProps<{
    suggestion: ResultSuggestion,
    chainId: ChainIDs,
    resultType: ResultType,
    barStyle: SearchbarStyle,
    barPurpose: SearchbarPurpose
}>()

function formatEmbeddedSubcategoryCell () : string {
  const label = SubCategoryInfo[TypeInfo[props.resultType].subCategory].title

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

function formatEmbeddedIdentificationCell () : string {
  if (isOutputAnAPIresponse(props.resultType, 'name') && !props.suggestion.nameWasUnknown) {
    return props.suggestion.output.name
  }
  return props.suggestion.output.lowLevelData
}
</script>

<template>
  <div
    v-if="barStyle == SearchbarStyle.Gaudy || barStyle == SearchbarStyle.Discreet"
    class="row-gaudyordiscreet"
    :class="barStyle"
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

  <div
    v-else-if="barStyle == SearchbarStyle.Embedded"
    class="row-embedded"
    :class="barStyle"
  >
    <div v-if="chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon not-alone" />
      <IconNetwork :chain-id="chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon alone" />
    </div>
    <div class="cell-subcategory" :class="barStyle">
      {{ formatEmbeddedSubcategoryCell() }}
    </div>
    <div class="cell-blockchaininfo-common cell-bi-identification" :class="barStyle">
      <BcSearchbarMiddleEllipsis :width-is-fixed="true">
        {{ formatEmbeddedIdentificationCell() }}
      </BcSearchbarMiddleEllipsis>
    </div>
    <div v-if="suggestion.output.description !== ''" class="cell-blockchaininfo-common cell-bi-description" :class="barStyle">
      {{ formatEmbeddedDescriptionCell() }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

@mixin common-to-both-rows {
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
    &.gaudy,
    &.embedded {
      background-color: var(--dropdown-background-hover);
    }
    &.discreet {
      background-color: var(--searchbar-background-hover-discreet);
    }
  }
  &:active {
    &.gaudy,
    &.embedded {
      background-color: var(--button-color-pressed);
    }
    &.discreet {
      background-color: var(--searchbar-background-pressed-discreet);
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

  .cell-name,
  .cell-subcategory {
    grid-column: 2;
    grid-row: 1;
    display: inline-block;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }
}

// specific style for the gaudy and discreet modes

.row-gaudyordiscreet {
  @include common-to-both-rows;

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
      &.greyish.gaudy {
        color: var(--searchbar-text-detail-gaudy);
      }
      &.greyish.discreet {
        color: var(--searchbar-text-detail-discreet);
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
      &.gaudy {
        color: var(--searchbar-text-detail-gaudy);
      }
      &.discreet {
        color: var(--searchbar-text-detail-discreet);
      }
    }
  }
}

// specific style for the embedded mode

.row-embedded {
  @include common-to-both-rows;

  @media (min-width: 600px) { // large screen
    grid-template-columns: 40px 106px auto min-content;
  }
  @media (max-width: 600px) { // mobile
    grid-template-columns: 40px auto min-content;
  }

  .cell-subcategory {
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

  .cell-bi-identification {
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
}
</style>
