<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faMagnifyingGlass, faPlus,
} from '@fortawesome/pro-solid-svg-icons'
import {
  type SearchbarColors,
  type SearchbarPurpose,
  SearchbarPurposeInfo,
  type SearchbarShape,
} from '~/types/searchbar'

defineProps<{
  barPurpose: SearchbarPurpose
  barShape: SearchbarShape
  colorTheme: SearchbarColors
}>()
</script>

<template>
  <Button
    v-if="SearchbarPurposeInfo[barPurpose].button === 'add'"
    class="p-button-icon-only plus"
  >
    <FontAwesomeIcon :icon="faPlus" />
  </Button>
  <span
    v-else-if="SearchbarPurposeInfo[barPurpose].button === 'search'"
    class="magnifier-button"
    :class="[barShape, colorTheme]"
  >
    <FontAwesomeIcon :icon="faMagnifyingGlass" />
  </span>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";

.plus {
  font-size: 18px;
}

.magnifier-button {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  border-radius: var(--border-radius);
  cursor: pointer;

  &.big {
    font-size: 18px;
    color: var(--text-color);
    background-color: transparent;
  }
  &.medium {
    font-size: 15px;
    color: var(--input-placeholder-text-color);
    background-color: transparent;
  }
  /*
  &:hover {
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
      background-color: var(--button-color-pressed);
    }
  } */
}
</style>
