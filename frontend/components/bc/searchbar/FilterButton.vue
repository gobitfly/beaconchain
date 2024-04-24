<script setup lang="ts">
import {
  SearchbarStyle
} from '~/types/searchbar'

defineProps<{
  barStyle : SearchbarStyle,
  state? : boolean,
  look? : 'on'|'off' // forces the look of the button statically instead of having the color changing with its state
}>()

const emit = defineEmits<{(e: 'change', activated : boolean) : void}>()
</script>

<template>
  <label class="frame" :class="[barStyle, look]">
    <input
      type="checkbox"
      class="hidden-checkbox"
      :true-value="true"
      :false-value="false"
      :checked="state"
      :onchange="(e:any) => {emit('change', e.target.checked)}"
    >
    <slot />
  </label>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.frame {
  display: inline-block;
  position: relative;
  box-sizing: border-box;
  cursor: pointer;
  border-radius: 10px;
  height: 20px;
  padding-top: 3px;
  padding-left: 8px;
  padding-right: 8px;
  text-align: center;
  transition: 0.2s;
  @include fonts.small_text_bold;
  white-space: nowrap;
  overflow: clip;

  .hidden-checkbox {
    display: none;
    width: 0;
    height: 0;
  }

  &:not(.off) {
    &.on,
    &:has(.hidden-checkbox:checked) {
      background-color: var(--button-color-active);
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }

  &:not(.on) {
    &.gaudy,
    &.embedded {
      background-color: var(--searchbar-filter-unselected-gaudy);
    }
    &.discreet {
      background-color: var(--light-grey);
    }
    &:hover {
      background-color: var(--light-grey-3);
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }

  &.gaudy,
  &.embedded {
    color: var(--primary-contrast-color);
  }
  &.discreet {
    color: var(--light-black);
  }
}
</style>
