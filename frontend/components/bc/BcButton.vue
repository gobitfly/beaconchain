<script lang="ts" setup>

const props =defineProps<{
  /**
   * ℹ️ should only be used rarely, e.g. in cases where the action should not be triggerd twice
   */
  isDisabled?: boolean,
  /**
   * ♿️ buttons that are aria-disabled are still perceivable by screen readers
   * as they can still be focused on 
   */
  isAriaDisabled?: boolean,
  variant?: 'secondary' // | 'red'
}>()

const shouldAppearDisabled = computed(() => props.isDisabled || props.isAriaDisabled)
</script>

<template> 
  <Button 
    type="button"
    :disabled="isDisabled"
    :aria-disabled="isAriaDisabled"
    :class="{
      'bc-button--secondary': !shouldAppearDisabled && variant === 'secondary',
      'bc-button--disabled': shouldAppearDisabled,
      // 'bc-button--red': variant === 'red'
    }"
  >
    <slot/>
    <span
      v-if="$slots.icon"
      class="bc-button__icon"
    >
      <slot name="icon"/>
    </span>
  </Button>

 </template>

<style lang="scss" scoped>
  .bc-button--secondary {
    border-color: var(--button-secondary-border-color);
    background-color: var(--button-secondary-background-color);
    color: var(--button-secondary-color);
    &:hover {
      background-color: var(--button-secondary-background-color--hover);
      border-color: var(--button-secondary-border-color);
    }
    &:active {
      background-color: var(--button-secondary-background-color);
      border-color: var(--button-secondary-border-color);
    }
  }
  .bc-button--disabled {
    &,
    &:hover,
    &:focus {
      background-color: var(--button-color-disabled);
      border-color: var(--button-color-disabled);
      color: var(--button-text-color-disabled);
      cursor: not-allowed;
    }
  }
  .bc-button__icon {
    margin-left: var(--padding-small);
  }
</style>