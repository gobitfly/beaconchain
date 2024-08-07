<script setup lang="ts">
import type {
  SearchbarColors,
  SearchbarDropdownLayout,
  SearchbarShape,
} from '~/types/searchbar'

defineProps<{
  barShape: SearchbarShape,
  colorTheme: SearchbarColors,
  dropdownLayout: SearchbarDropdownLayout,
  look?: 'off' | 'on', // forces the look of the button statically instead of having the color changing with its state
  state?: boolean,
}>()

const emit = defineEmits<{ (e: 'change', activated: boolean): void }>()
</script>

<template>
  <label
    class="frame"
    :class="[barShape, colorTheme, dropdownLayout, look]"
  >
    <input
      type="checkbox"
      class="hidden-checkbox"
      :true-value="true"
      :false-value="false"
      :checked="state"
      :onchange="
        (e: any) => {
          emit('change', e.target.checked);
        }
      "
    >
    <div class="content">
      <slot />
    </div>
  </label>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.frame {
  display: inline-flex;
  position: relative;
  box-sizing: border-box;
  cursor: pointer;
  user-select: none;
  border-radius: 10px;
  height: 22px;
  padding-left: 8px;
  padding-right: 8px;
  text-align: center;
  transition: 0.2s;
  white-space: nowrap;
  overflow: clip;
  @include fonts.small_text_bold;
  @media (pointer: coarse) {
    border-radius: 15px;
    height: 30px;
  }
  &.narrow-dropdown {
    letter-spacing: -0.02em;
  }
  &.default {
    border: 1px solid var(--container-border-color);
  }

  .hidden-checkbox {
    display: none;
    width: 0;
    height: 0;
  }

  &:not(.off) {
    &.on,
    &:has(.hidden-checkbox:checked) {
      background-color: var(--button-color-active);
      &.default {
        border: 1px solid var(--button-color-active);
      }
      &:hover {
        @media (hover: hover) {
          background-color: var(--button-color-hover);
        }
      }
      &:active {
        background-color: var(--button-color-pressed);
      }
    }
  }

  &:not(.on) {
    &.darkblue,
    &.lightblue {
      background-color: var(--light-grey);
    }
    @media (hover: hover) {
      &:hover {
        &.default {
          background-color: var(--container-border-color);
        }
        &.darkblue,
        &.lightblue {
          background-color: var(--light-grey-3);
        }
      }
    }
    &:active {
      background-color: var(--button-color-pressed);
    }
  }

  &.default {
    color: var(--text-color);
  }
  &.darkblue,
  &.lightblue {
    color: var(--light-black);
  }
}

.content {
  display: inline-flex;
  position: relative;
  margin-top: auto;
  margin-bottom: auto;
}
</style>
