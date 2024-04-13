<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faMagnifyingGlass, faPlus } from '@fortawesome/pro-solid-svg-icons'
import {
  SearchbarStyle,
  SearchbarPurpose,
  SearchbarPurposeInfo
} from '~/types/searchbar'

defineProps<{
    barStyle: SearchbarStyle,
    barPurpose: SearchbarPurpose
}>()

</script>

<template>
  <Button v-if="SearchbarPurposeInfo[barPurpose].button === 'add'" class="p-button-icon-only plus">
    <FontAwesomeIcon :icon="faPlus" />
  </Button>
  <span v-else-if="SearchbarPurposeInfo[barPurpose].button === 'search'" class="magnifier-button" :class="barStyle">
    <FontAwesomeIcon :icon="faMagnifyingGlass" />
  </span>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.plus {
  font-size: 18px;
}

.magnifier-button {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  border-radius: var(--border-radius);
  cursor: pointer;

  &.gaudy {
    font-size: 18px;
    color: var(--text-color);
    background-color: transparent;
    &:hover {
      background-color: var(--dropdown-background-hover);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }
  &.discreet {
    font-size: 15px;
    color: var(--input-placeholder-text-color);
    background-color: transparent;
    &:hover {
      background-color: var(--searchbar-background-hover-discreet);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }
}
</style>
