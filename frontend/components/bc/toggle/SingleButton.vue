<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
interface Props {
  icon?: IconDefinition,
  text?: string
}
const props = defineProps<Props>()
const selected = defineModel<boolean | undefined>({ required: true })
</script>

<template>
  <ToggleButton v-model="selected" class="bc-toggle" :on-label="props.text" :off-label="props.text">
    <template #icon="slotProps">
      <slot name="icon" v-bind="slotProps">
        <FontAwesomeIcon v-if="props.icon" :icon="props.icon" />
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
        width: 36px;
        height: 36px;
      }
    }
  }
}
</style>
