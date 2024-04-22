<script setup lang="ts">
import {
  SearchbarStyle
} from '~/types/searchbar'

const props = defineProps<{
  barStyle : SearchbarStyle,
  forcedColor? : number // to controle the color from the parent instead of having the color matching the real state of the button. 0 = deactivated, any other number = activated.
}>()
const activated = defineModel<boolean>()
const emit = defineEmits<{(e: 'change', activated : boolean) : void}>()

const classForcingColorTheme = computed(() => props.forcedColor === undefined ? '' : (props.forcedColor ? 'forced-on' : 'forced-off'))
</script>

<template>
  <span class="frame">
    <label class="button" :class="classForcingColorTheme">
      <input
        type="checkbox"
        class="hidden-checkbox"
        :true-value="true"
        :false-value="false"
        :checked="activated"
        :onchange="(e:any) => emit('change', e.target.checked)"
      >
      <span class="face" :class="[barStyle, classForcingColorTheme]">
        <slot />
      </span>
    </label>
  </span>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.frame {
  display: inline-block;
  position: relative;
  .button {
    @include fonts.small_text_bold;

    .hidden-checkbox {
      display: none;
      width: 0;
      height: 0;
    }

    .hidden-checkbox:checked + .face {
      background-color: var(--button-color-active);
      &:hover {
        background-color: var(--button-color-hover);
      }
      &:active {
        background-color: var(--button-color-pressed);
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

      &.gaudy,
      &.gaudy.forced-off {
        color: var(--primary-contrast-color);
        background-color: var(--searchbar-filter-unselected-gaudy);
      }
      &.discreet,
      &.discreet.forced-off {
        color: var(--light-black);
        background-color: var(--light-grey);
      }

      &:hover,
      &:hover.forced-off {
        background-color: var(--light-grey-3);
      }
      &:active,
      &:active.forced-off {
        background-color: var(--button-color-pressed);
      }

      &.forced-on {
        background-color: var(--button-color-active);
        &:hover {
          background-color: var(--button-color-hover);
        }
        &:active {
          background-color: var(--button-color-pressed);
        }
      }
    }
  }
}
</style>
