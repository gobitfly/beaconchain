<script setup lang="ts">
/*
 * If you want to change the behavior of the component or the information it displays, it is possible that you simply need to change a few parameters
 * in searchbar.ts rather than altering the code of the component. The possibilities offered by this configuration file are explanined in readme.md
 */
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
  SearchbarPurposeInfo,
  SuggestionrowCells,
  getI18nPathOfTranslatableLitteral
} from '~/types/searchbar'

const props = defineProps<{
    suggestion: ResultSuggestion,
    barStyle: SearchbarStyle,
    barPurpose: SearchbarPurpose
}>()

const { t } = useI18n()

function formatSubcategoryCell () : string {
  const i18nPathOfSubcategoryTitle = getI18nPathOfTranslatableLitteral(SubCategoryInfo[TypeInfo[props.suggestion.type].subCategory].title)
  let label = t(i18nPathOfSubcategoryTitle, props.suggestion.count)

  if (props.suggestion.count >= 2) {
    label = String(props.suggestion.count) + ' ' + label
  }
  return label
}

function formatIdentificationCell () : string {
  if (wasOutputDataGivenByTheAPI(props.suggestion.type, 'name') && !props.suggestion.nameWasUnknown) {
    return props.suggestion.output.name
  }
  return props.suggestion.output.lowLevelData
}

function formatDescriptionCell () : string {
  if (wasOutputDataGivenByTheAPI(props.suggestion.type, 'description')) {
    // we tell the user what is the data that they see (ex: "Index" for a validator index)
    switch (props.suggestion.type) {
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
    v-if="SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.NameDescriptionLowlevelCategory"
    class="rowstyle_name-description-low-level-category"
    :class="barStyle"
  >
    <!-- In this mode, all possible cells are shown (as originally designed on Figma) -->
    <div v-if="props.suggestion.chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="props.suggestion.type" class="type-icon not-alone" />
      <IconNetwork :chain-id="props.suggestion.chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="props.suggestion.type" class="type-icon alone" />
    </div>
    <BcSearchbarMiddleEllipsis
      class="cell_name"
      :class="barStyle"
      :text="suggestion.output.name"
    />
    <BcSearchbarMiddleEllipsis class="group_blockchain-info" :class="barStyle">
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.description !== ''"
        :text="suggestion.output.description"
        :initial-flex-grow="1"
        class="cell_bi_description"
        :class="barStyle"
      />
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.lowLevelData !== ''"
        :text="suggestion.output.lowLevelData"
        class="cell_bi_low-level-data"
        :class="[barStyle,(suggestion.output.description !== '')?'greyish':'']"
      />
    </BcSearchbarMiddleEllipsis>
    <div class="cell-category" :class="barStyle">
      <span class="category-label" :class="barStyle">
        {{ t(...CategoryInfo[TypeInfo[props.suggestion.type].category].title) }}
      </span>
    </div>
  </div>

  <div
    v-else-if="SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.SubcategoryIdentificationDescription"
    class="rowstyle_subcategory-identification-description"
    :class="barStyle"
  >
    <!-- In this mode, we show less cells and their content comes from dedicated functions instead of a pure copy of `props.suggestion.output` -->
    <div v-if="props.suggestion.chainId !== ChainIDs.Any" class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="props.suggestion.type" class="type-icon not-alone" />
      <IconNetwork :chain-id="props.suggestion.chainId" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </div>
    <div v-else class="cell-icons" :class="barStyle">
      <BcSearchbarTypeIcons :type="props.suggestion.type" class="type-icon alone" />
    </div>
    <div class="cell-subcategory" :class="barStyle">
      {{ formatSubcategoryCell() }}
    </div>
    <BcSearchbarMiddleEllipsis
      class="cells_blockchain-info_common cell_bi_identification"
      :class="barStyle"
      :text="formatIdentificationCell()"
    />
    <div v-if="suggestion.output.description !== ''" class="cells_blockchain-info_common cell_bi_description" :class="barStyle">
      {{ formatDescriptionCell() }}
    </div>
  </div>

  <!-- If you want to show other cells or change their format, it might be good to implement a new mode here instead of modiying the modes above.
       To make the bar use your new mode, add its name into the `SuggestionrowCells` enum in `searchbar.ts`, and update the `SearchbarPurposeInfo` record there. -->
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

@mixin common-to-all-rowstyles {
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

  .cell_name,
  .cell-subcategory {
    grid-column: 2;
    grid-row: 1;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }
}

// specific style when SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.NameDescriptionLowlevelCategory

.rowstyle_name-description-low-level-category {
  @include common-to-all-rowstyles;

  @media (min-width: 600px) { // large screen
    &.gaudy,
    &.embedded {
      grid-template-columns: 40px 106px auto 114px;
    }
    &.discreet {
      grid-template-columns: 40px 114px auto;
    }
  }
  @media (max-width: 600px) { // mobile
    grid-template-columns: 40px 114px auto;
  }

  .cell_name {
    font-weight: var(--roboto-medium);
    margin-right: 16px;
  }

  .group_blockchain-info {
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
    white-space: nowrap;  // makes sure that the two spans (description + lowleveldata) stay on the same line

    .cell_bi_description {
      position: relative;
      @media (min-width: 600px) { // large screen
        &.gaudy,
        &.embedded {
          margin-right: 0.5em;
        }
      }
    }

    .cell_bi_low-level-data {
      position: relative;
      flex-grow: 1;
      text-align: justify;
      text-justify: inter-character;
      &.greyish {
        &.gaudy,
        &.embedded {
          color: var(--searchbar-text-detail-gaudy);
        }
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
      &.gaudy,
      &.embedded {
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
      &.gaudy,
      &.embedded {
        color: var(--searchbar-text-detail-gaudy);
      }
      &.discreet {
        color: var(--searchbar-text-detail-discreet);
      }
    }
  }
}

// specific style when SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.SubcategoryIdentificationDescription

.rowstyle_subcategory-identification-description {
  @include common-to-all-rowstyles;

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

  .cells_blockchain-info_common {
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }

  .cell_bi_identification {
    @media (min-width: 600px) { // large screen
      grid-column: 3;
      font-weight: var(--roboto-medium);
    }
    @media (max-width: 600px) { // mobile
      grid-row: 2;
      grid-column-end: span 2;
      font-weight: var(--roboto-regular);
    }
    text-align: justify;
    text-justify: inter-character;
  }

  .cell_bi_description {
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
