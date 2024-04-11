<script setup lang="ts">
import { ChainIDs } from '~/types/networks'
import {
  CategoryInfo,
  SubCategoryInfo,
  TypeInfo,
  ResultType,
  wasOutputDataGivenByTheAPI,
  type ResultSuggestion,
  SearchbarStyle,
  SearchbarPurpose,
  getI18nPathOfTranslatableLitteral
} from '~/types/searchbar'

const props = defineProps<{
    suggestion: ResultSuggestion,
    chainId: ChainIDs,
    resultType: ResultType,
    barStyle: SearchbarStyle,
    barPurpose: SearchbarPurpose
}>()

const { t } = useI18n()

function formatEmbeddedSubcategoryCell () : string {
  const i18nPathOfSubcategoryTitle = getI18nPathOfTranslatableLitteral(SubCategoryInfo[TypeInfo[props.resultType].subCategory].title)
  let label = t(i18nPathOfSubcategoryTitle, props.suggestion.count)

  if (props.suggestion.count >= 2) {
    label = String(props.suggestion.count) + ' ' + label
  }
  return label
}

function formatEmbeddedIdentificationCell () : string {
  if (wasOutputDataGivenByTheAPI(props.resultType, 'name') && !props.suggestion.nameWasUnknown) {
    return props.suggestion.output.name
  }
  return props.suggestion.output.lowLevelData
}

function formatEmbeddedDescriptionCell () : string {
  if (wasOutputDataGivenByTheAPI(props.resultType, 'description')) {
    // we tell the user what is the data that they see (ex: "Index" for a validator index)
    switch (props.resultType) {
      case ResultType.ValidatorsByIndex :
      case ResultType.ValidatorsByPubkey :
        return t('common.index') + ' ' + props.suggestion.output.description
      // more cases might arise in the future
    }
  }
  return props.suggestion.output.description
}
</script>

<template>
  <div
    v-if="barStyle == SearchbarStyle.Gaudy || barStyle == SearchbarStyle.Discreet"
    class="rowstyle-gaudyordiscreet"
    :class="barStyle"
  >
    <div v-if="chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon not-alone" />
      <IconNetwork :chain-id="chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="resultType" class="type-icon alone" />
    </div>
    <BcSearchbarMiddleEllipsis
      class="cell-name"
      :class="barStyle"
      :text="suggestion.output.name"
    />
    <BcSearchbarMiddleEllipsis class="group-blockchaininfo" :class="barStyle">
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.description !== ''"
        :text="suggestion.output.description"
        :dont-clip-under="16"
        :max-flex-grow="1"
        class="cell-bi-description"
        :class="barStyle"
      />
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.lowLevelData !== ''"
        :text="suggestion.output.lowLevelData"
        class="cell-bi-lowleveldata"
        :class="[barStyle,(suggestion.output.description !== '')?'greyish':'']"
      />
    </BcSearchbarMiddleEllipsis>
    <div class="cell-category" :class="barStyle">
      <span class="category-label" :class="barStyle">
        {{ t(...CategoryInfo[TypeInfo[resultType].category].title) }}
      </span>
    </div>
  </div>

  <div
    v-else-if="barStyle == SearchbarStyle.Embedded"
    class="rowstyle-embedded"
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
    <BcSearchbarMiddleEllipsis
      class="cells-blockchaininfo-common cell-bi-identification"
      :class="barStyle"
      :text="formatEmbeddedIdentificationCell()"
    />
    <div v-if="suggestion.output.description !== ''" class="cells-blockchaininfo-common cell-bi-description" :class="barStyle">
      {{ formatEmbeddedDescriptionCell() }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

@mixin common-to-both-rowstyles {
  cursor: pointer;
  display: grid;
  position: relative;
  right: 0px;
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
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }
}

// specific style for the gaudy and discreet modes

.rowstyle-gaudyordiscreet {
  @include common-to-both-rowstyles;

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
    margin-right: 16px;
  }

  .group-blockchaininfo {
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
    white-space: nowrap;  // this has an effect on a large screen in gaudy mode only, it makes sure that the two spans (description + lowleveldata) stay on the same line

    .cell-bi-description {
      position: relative;
      @media (min-width: 600px) { // large screen
        &.gaudy {
          margin-right: 0.5em;
        }
      }
    }

    .cell-bi-lowleveldata {
      position: relative;
      flex-grow: 1;
      text-align: justify;
      text-justify: inter-character;
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

.rowstyle-embedded {
  @include common-to-both-rowstyles;

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

  .cells-blockchaininfo-common {
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }

  .cell-bi-identification {
    @media (min-width: 600px) { // large screen
      grid-column: 3;
      font-weight: var(--roboto-medium);
    }
    @media (max-width: 600px) { // mobile
      grid-row: 2;
      grid-column-end: span 2;
      font-weight: var(--roboto-regular);
    }
    width: 100%;
    text-align: justify;
    text-justify: inter-character;
  }

  .cell-bi-description {
    @media (min-width: 600px) { // large screen
      grid-column: 4;
    }
    @media (max-width: 600px) { // mobile
      grid-row: 1;
      grid-column: 3;
      color: var(--searchbar-text-detail-gaudy);
    }
    width: 100px;
    margin-left: auto;
    justify-content: right;
  }
}
</style>
