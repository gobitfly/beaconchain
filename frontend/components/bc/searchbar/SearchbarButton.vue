<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faMagnifyingGlass, faPlus } from '@fortawesome/pro-solid-svg-icons'
import {
  type SearchBarStyle,
  SearchBarPurpose
} from '~/types/searchbar'

defineProps<{
    barStyle: SearchBarStyle,
    barPurpose: SearchBarPurpose
}>()

</script>

<template>
  <Button v-if="barPurpose == SearchBarPurpose.Accounts || barPurpose == SearchBarPurpose.Validators" class="p-button-icon-only plus">
    <FontAwesomeIcon :icon="faPlus" />
  </Button>
  <span v-else class="magnifier-button" :class="barStyle">
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
}
</style>
