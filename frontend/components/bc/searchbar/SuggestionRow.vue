<script setup lang="ts">
/*
 * If you want to change the behavior of the component or the information it displays,
 it is possible that you simply need to change a few parameters
 * in searchbar.ts rather than altering the code of the component.
 The possibilities offered by this configuration file are explanined in readme.md
 */
import { ChainIDs } from '~/types/network'
import {
  CategoryInfo,
  getI18nPathOfTranslatableLitteral,
  type ResultSuggestionInternal,
  ResultType,
  type SearchbarColors,
  type SearchbarDropdownLayout,
  type SearchbarPurpose,
  SearchbarPurposeInfo,
  type SearchbarShape,
  SubCategoryInfo,
  SuggestionrowCells,
  TypeInfo,
  wasOutputDataGivenByTheAPI,
} from '~/types/searchbar'

const props = defineProps<{
  barPurpose: SearchbarPurpose,
  barShape: SearchbarShape,
  colorTheme: SearchbarColors,
  dropdownLayout: SearchbarDropdownLayout,
  screenWidthCausingSuddenChange: number,
  suggestion: ResultSuggestionInternal,
}>()

const { t } = useTranslation()

function formatSubcategoryCell(): string {
  const i18nPathOfSubcategoryTitle = getI18nPathOfTranslatableLitteral(
    SubCategoryInfo[TypeInfo[props.suggestion.type].subCategory].title,
  )
  let label = t(i18nPathOfSubcategoryTitle, props.suggestion.count)

  if (props.suggestion.count >= 2) {
    label = String(props.suggestion.count) + ' ' + label
  }
  return label
}

function formatIdentificationCell(): string {
  if (
    wasOutputDataGivenByTheAPI(props.suggestion.type, 'name')
    && !props.suggestion.nameWasUnknown
  ) {
    return props.suggestion.output.name
  }
  return props.suggestion.output.lowLevelData
}

function formatDescriptionCell(): string {
  if (wasOutputDataGivenByTheAPI(props.suggestion.type, 'description')) {
    // we tell the user what is the data that they see (ex: "Index" for a validator index)
    switch (props.suggestion.type) {
      case ResultType.ValidatorsByIndex:
      case ResultType.ValidatorsByPubkey:
        return t('common.index') + ' ' + props.suggestion.output.description
      // more cases might arise in the future
    }
  }
  return props.suggestion.output.description
}

const deactivationClass = props.suggestion.lacksPremiumSubscription
  ? 'deactivated'
  : ''
</script>

<template>
  <div
    v-if="
      SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows
        === SuggestionrowCells.NameDescriptionLowlevelCategory
    "
    class="rowstyle_name-description-low-level-category"
    :class="[barShape, colorTheme, dropdownLayout, deactivationClass]"
  >
    <!-- In this mode, all possible cells are shown (this was the very first design on Figma) -->
    <div
      v-if="props.suggestion.chainId !== ChainIDs.Any"
      class="cell-icons"
      :class="[barShape, dropdownLayout, deactivationClass]"
    >
      <BcSearchbarTypeIcons
        :type="props.suggestion.type"
        class="type-icon not-alone"
      />
      <IconNetwork
        :chain-id="props.suggestion.chainId"
        :colored="true"
        :harmonize-perceived-size="true"
        :do-not-adapt-to-color-theme="colorTheme !== 'default'"
        class="network-icon"
      />
    </div>
    <div
      v-else
      class="cell-icons"
      :class="[barShape, dropdownLayout, deactivationClass]"
    >
      <BcSearchbarTypeIcons
        :type="props.suggestion.type"
        class="type-icon alone"
      />
    </div>
    <BcSearchbarMiddleEllipsis
      class="cell_name"
      :class="[barShape, dropdownLayout, deactivationClass]"
      :text="suggestion.output.name"
      :width-mediaquery-threshold="screenWidthCausingSuddenChange"
    />
    <BcSearchbarMiddleEllipsis
      class="group_blockchain-info"
      :class="[barShape, dropdownLayout, deactivationClass]"
      :width-mediaquery-threshold="screenWidthCausingSuddenChange"
    >
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.description !== ''"
        :text="suggestion.output.description"
        :initial-flex-grow="1"
        class="cell_bi_description"
        :class="[barShape, colorTheme, dropdownLayout]"
      />
      <BcSearchbarMiddleEllipsis
        v-if="suggestion.output.lowLevelData !== ''"
        :text="suggestion.output.lowLevelData"
        class="cell_bi_low-level-data"
        :class="[
          barShape,
          colorTheme,
          dropdownLayout,
          suggestion.output.description ? 'greyish' : '',
        ]"
      />
    </BcSearchbarMiddleEllipsis>
    <div
      class="premium-invitation"
      :class="dropdownLayout"
      @click="(e: Event) => e.stopPropagation()"
    >
      <BcPremiumGem
        v-if="suggestion.lacksPremiumSubscription"
        class="gem"
      />
    </div>
    <div
      class="cell-category"
      :class="[barShape, dropdownLayout, deactivationClass]"
    >
      <span
        class="category-label"
        :class="[barShape, colorTheme, dropdownLayout]"
      >
        {{ t(...CategoryInfo[TypeInfo[props.suggestion.type].category].title) }}
      </span>
    </div>
  </div>

  <div
    v-else-if="
      SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows
        === SuggestionrowCells.SubcategoryIdentificationDescription
    "
    class="rowstyle_subcategory-identification-description"
    :class="[barShape, colorTheme, dropdownLayout, deactivationClass]"
  >
    <!-- In this mode, we show less cells and their content comes from dedicated functions instead of a pure copy of
     `props.suggestion.output` -->
    <div
      v-if="props.suggestion.chainId !== ChainIDs.Any"
      class="cell-icons"
      :class="[barShape, deactivationClass]"
    >
      <BcSearchbarTypeIcons
        :type="props.suggestion.type"
        class="type-icon not-alone"
      />
      <IconNetwork
        :chain-id="props.suggestion.chainId"
        :colored="true"
        :harmonize-perceived-size="true"
        :do-not-adapt-to-color-theme="colorTheme !== 'default'"
        class="network-icon"
      />
    </div>
    <div
      v-else
      class="cell-icons"
      :class="[barShape, deactivationClass]"
    >
      <BcSearchbarTypeIcons
        :type="props.suggestion.type"
        class="type-icon alone"
      />
    </div>
    <div
      class="cell-subcategory"
      :class="[barShape, dropdownLayout, deactivationClass]"
    >
      {{ formatSubcategoryCell() }}
    </div>
    <BcSearchbarMiddleEllipsis
      class="cell_bi_identification"
      :class="[barShape, dropdownLayout, deactivationClass]"
      :text="formatIdentificationCell()"
      :width-mediaquery-threshold="screenWidthCausingSuddenChange"
    />
    <div
      class="premium-invitation"
      :class="dropdownLayout"
      @click="(e: Event) => e.stopPropagation()"
    >
      <BcPremiumGem
        v-if="suggestion.lacksPremiumSubscription"
        class="gem"
      />
    </div>
    <div
      v-if="suggestion.output.description !== ''"
      class="cell_bi_description"
      :class="[barShape, colorTheme, dropdownLayout, deactivationClass]"
    >
      {{ formatDescriptionCell() }}
    </div>
  </div>

  <!-- If you want to show other cells or change their format, it might be good to implement a new mode here instead of
   modiying the modes above.
   To make the bar use your new mode, add its name into the `SuggestionrowCells` enum in `searchbar.ts`,
   and update the `SearchbarPurposeInfo` record there. -->
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/fonts.scss";

.deactivated {
  opacity: 0.6;
  pointer-events: none;
}

@mixin common-to-all-rowstyles {
  cursor: pointer;
  user-select: none;
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
  @include fonts.standard_text;

  &.deactivated {
    opacity: unset;
  }

  &:hover:not(.deactivated) {
    &.default {
      background-color: var(--dropdown-background-hover);
    }
    &.darkblue {
      background-color: var(--searchbar-background-hover-darkblue);
    }
    &.lightblue {
      background-color: var(--searchbar-background-hover-lightblue);
    }
  }
  &:active {
    &.default {
      background-color: var(--button-color-pressed);
    }
    &.darkblue,
    &.lightblue {
      background-color: var(--searchbar-background-pressed-blue);
    }
  }

  .cell-icons {
    position: relative;
    grid-column: 1;
    grid-row: 1;
    display: flex;
    margin-top: auto;
    margin-bottom: auto;
    width: 30px;
    height: 36px;

    &.narrow-dropdown {
      grid-row-end: span 2;
    }
    .type-icon {
      width: 20px;
      max-height: 20px;
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

  .premium-invitation {
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
    justify-content: right;
    pointer-events: auto;
    .gem {
      margin-left: 10px;
      margin-right: 10px;
    }
  }
}

// specific style when SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.NameDescriptionLowlevelCategory

.rowstyle_name-description-low-level-category {
  @include common-to-all-rowstyles;

  &.large-dropdown {
    grid-template-columns: 40px 130px 1fr min-content 130px;
  }
  &.narrow-dropdown {
    grid-template-columns: 40px 130px 1fr min-content;
  }

  .cell_name {
    font-weight: var(--standard_text_medium_font_weight);
    margin-right: 16px;
  }

  .group_blockchain-info {
    grid-column: 3;
    grid-row: 1;
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
    font-weight: var(--standard_text_medium_font_weight);
    white-space: nowrap; // makes sure that the two spans (description + lowleveldata) stay on the same line
    &.narrow-dropdown {
      grid-row-end: span 2;
      flex-direction: column;
    }

    .cell_bi_description {
      position: relative;
      &.large-dropdown {
        margin-right: 0.5em;
      }
    }

    .cell_bi_low-level-data {
      position: relative;
      flex-grow: 1;
      &.greyish {
        &.default {
          color: var(--searchbar-text-detail-default);
        }
      }
      &.greyish.darkblue {
        color: var(--searchbar-text-detail-darkblue);
      }
      &.greyish.lightblue {
        color: var(--searchbar-text-detail-lightblue);
      }
    }
  }

  .premium-invitation {
    grid-column: 4;
    &.narrow-dropdown {
      grid-row: 1;
      grid-row-end: span 2;
    }
  }

  .cell-category {
    display: flex;
    position: relative;
    &.large-dropdown {
      grid-column: 5;
      grid-row: 1;
      margin-top: auto;
      margin-bottom: auto;
      margin-left: 16px;
    }
    &.narrow-dropdown {
      grid-column: 2;
      grid-row: 2;
    }

    .category-label {
      display: inline-flex;
      position: relative;
      &.large-dropdown {
        margin-left: auto;
      }
      &.default {
        color: var(--searchbar-text-detail-default);
      }
      &.darkblue {
        color: var(--searchbar-text-detail-darkblue);
      }
      &.lightblue {
        color: var(--searchbar-text-detail-lightblue);
      }
    }
  }
}

// specific style when SearchbarPurposeInfo[barPurpose].cellsInSuggestionRows === SuggestionrowCells.SubcategoryIdentificationDescription

.rowstyle_subcategory-identification-description {
  @include common-to-all-rowstyles;

  &.large-dropdown {
    grid-template-columns: 40px minmax(126px, min-content) 1fr min-content min-content;
  }
  &.narrow-dropdown {
    grid-template-columns: 40px 1fr min-content min-content;
  }

  .cell-subcategory {
    box-sizing: border-box;
    margin-right: auto;
    white-space: nowrap;
    &.large-dropdown {
      font-weight: var(--standard_text_medium_font_weight);
      padding-right: 16px;
    }
    &.narrow-dropdown {
      font-weight: var(--standard_text_font_weight);
    }
  }

  @mixin cells_blockchain-info_common {
    display: flex;
    position: relative;
    margin-top: auto;
    margin-bottom: auto;
  }

  .cell_bi_identification {
    @include cells_blockchain-info_common;
    &.large-dropdown {
      grid-column: 3;
      font-weight: var(--standard_text_medium_font_weight);
    }
    &.narrow-dropdown {
      grid-row: 2;
      grid-column: 2;
      grid-column-end: span 3;
      font-weight: var(--standard_text_font_weight);
    }
  }

  .premium-invitation {
    &.large-dropdown {
      grid-column: 4;
    }
    &.narrow-dropdown {
      grid-row: 1;
      grid-column: 3;
    }
  }

  .cell_bi_description {
    @include cells_blockchain-info_common;

    box-sizing: border-box;
    margin-left: auto;
    justify-content: right;
    white-space: nowrap;
    &.large-dropdown {
      grid-column: 5;
      width: 128px;
      padding-left: 16px;
    }
    &.narrow-dropdown {
      grid-row: 1;
      grid-column: 4;
      color: var(--searchbar-text-detail-default);
    }
  }
}
</style>
