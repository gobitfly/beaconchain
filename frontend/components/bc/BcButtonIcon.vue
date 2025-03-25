<script lang="ts" setup>
import type { Icon } from '~/components/bc/icon/BcIcon.vue'
import type { Rotation } from '~/components/BcRotate.vue'

const { variant = 'plain' } = defineProps<{
  isDisabled?: boolean,
  name: Icon,
  rotation?: Rotation,
  /**
   *
   * ♿️ screenreader text
   * every button with just an icon has to describe what it does
   */
  screenreaderText: TranslationInput,
  variant?: 'flat' | 'plain',
}>()
</script>

<template>
  <button
    class="bc-button-icon"
    :class="{
      'bc-button-icon--plain': variant === 'plain',
      'bc-button-icon--flat': variant === 'flat',
    }"
    :disabled="isDisabled"
  >
    <BcScreenreaderOnly
      :screenreader-text
    />
    <BcIcon
      v-if="name"
      :name
      :rotation
      :size="variant === 'flat' ? 'sm' : undefined"
    />
  </button>
</template>

<style lang="scss" scoped>
.bc-button-icon {
  --outline-width: 0.125rem;
  --outline-offset: 0.125rem;
  margin: calc(var(--outline-width) + var(--outline-offset));
  border-radius: 2px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: currentColor;

  transition: background-color 0.2s, color 0.2s, border-color 0.2s;

  &:focus-visible {
    outline: var(--outline-width) solid var(--blue-500);
    outline-offset: var(--outline-offset);
  }
}

.bc-button-icon--flat {
  height: 1.875rem;
  padding: var(--padding-small) var(--padding-small);
  border: 1px solid var(--container-border-color);
  background-color: var(--list-background);
  color: var(--button-secondary-color);

  &:hover {
    background: var(--list-hover-background);
    color: var(--list-hover-color);
  }
}
.bc-button-icon--plain {
  padding: 0;
  background-color: unset;
  border: unset;
}
</style>
