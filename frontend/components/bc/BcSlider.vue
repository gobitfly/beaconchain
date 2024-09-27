<script setup lang="ts">
const value = defineModel<number>({ required: true })
defineProps<{
  max?: number,
  min?: number,
  step?: number,
}>()
</script>

<template>
  <input
    v-model.number="value"
    :min
    :max
    :step
    class="bc-slider"
    type="range"
    v-bind="$attrs"
  >
</template>

<style scoped lang="scss">
  $heightTrack: 0.25rem;
  $backgroundColor: var(--primary-orange);
  $backgroundColor--disabled: var(--grey);
  $borderRadius: 1rem;
  $border: 0.1px solid $backgroundColor;
  $backgroundColorThumb: var(--light-grey);
  $thumbSize: 1.25rem;

  .bc-slider {
    width: 100%;
    appearance: none;
    background: transparent;
    cursor: pointer;

    &:focus-visible {
      outline-style: solid;
      outline-offset: 3px;
      border-radius: 2px;
    }

    &::-webkit-slider-runnable-track {
      background: $backgroundColor;
      height: $heightTrack;
      border-radius: $borderRadius;
      border: $border;
      display: inline-flex;
    }
     &[disabled]::-webkit-slider-runnable-track{
      background: $backgroundColor--disabled;
      border-color: $backgroundColor--disabled
    }

    &::-moz-range-track {
      background: $backgroundColor;
      height: $heightTrack;
      border-radius: $borderRadius;
    }
    &[disabled]::-moz-range-track {
      background: $backgroundColor--disabled;
    }

    $opticalCorrection: 1px;
    &::-webkit-slider-thumb {
      appearance: none;
      background-color: $backgroundColorThumb;
      width: $thumbSize;
      height: $thumbSize;
      border-radius: 50%;
      border: 2px solid $backgroundColor;
      margin-top: calc($heightTrack / 2 - $thumbSize / 2 - $opticalCorrection);
    }
    &[disabled]::-webkit-slider-thumb {
      border-color: $backgroundColor--disabled;
    }

    &::-moz-range-thumb {
      $thumbSize: 1rem;
      appearance: none;
      background-color: $backgroundColorThumb;
      width: $thumbSize;
      height: $thumbSize;
      border-radius: 50%;
      border: 2px solid $backgroundColor;
    }
    &[disabled]::-moz-range-thumb {
      border-color: $backgroundColor--disabled;
    }
  }
</style>
