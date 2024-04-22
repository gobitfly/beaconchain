<script setup lang="ts">
import {
  SearchbarStyle
} from '~/types/searchbar'

defineProps<{
  barStyle: SearchbarStyle
}>()
const activated = defineModel<boolean>()
const emit = defineEmits<{(e: 'change', activated : boolean) : void}>()
</script>

<template>
  <label class="frame">
    <input
      type="checkbox"
      class="hidden-checkbox"
      :true-value="true"
      :false-value="false"
      :checked="activated"
      :onchange="(e:any) => emit('change', e.target.checked)"
    >
    <span class="face" :class="barStyle">
      <slot />
    </span>
  </label>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.frame {
  @include fonts.small_text_bold;

  .hidden-checkbox {
    display: none;
    width: 0;
    height: 0;
    &:checked + .face {
      background-color: var(--button-color-active);
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }

  .face {
    display: inline-block;
    cursor: pointer;
    border-radius: 10px;
    height: 17px;
    padding-top: 2.5px;
    padding-left: 8px;
    padding-right: 8px;
    text-align: center;
    transition: 0.2s;

    &.gaudy {
      color: var(--primary-contrast-color);
      background-color: var(--searchbar-filter-unselected-gaudy);
    }
    &.discreet {
      color: var(--light-black);
      background-color: var(--light-grey);
    }

    &:hover {
      background-color: var(--light-grey-3);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }
}
</style>
