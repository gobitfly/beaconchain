<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
interface Props {
  icon?: IconDefinition,
  text?: string
  selected: boolean,
  disabled?:boolean,
}
defineProps<Props>()

</script>

<template>
  <ToggleButton class="bc-toggle" :disabled="disabled" :model-value="selected" :on-label="text" :off-label="text">
    <template #icon="slotProps">
      <slot name="icon" v-bind="slotProps">
        <FontAwesomeIcon v-if="icon" :icon="icon" />
      </slot>
    </template>
  </ToggleButton>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.bc-toggle {
  &.p-button {
    &.p-togglebutton {
      display: flex;
      flex-grow: 1;
      flex-direction: column;
      gap: 11px;

      width: 100%;
      height: 100%;
      padding: 16px 0 15px 0;
      border: 1px var(--container-border-color) solid;
      border-radius: var(--border-radius);
      background-color: var(--container-background);
      color: var(--text-color);

      &.p-highlight {
        border-color: var(--button-color-active);
        color: var(--button-color-active);
      }

      :deep(.p-button-label) {
        @include fonts.subtitle_text;
      }

      :deep(svg) {
        max-width: 36px;
      }
      &.p-disabled{
        opacity: 0.5;
        cursor: default;
      }
    }
  }
}
</style>
