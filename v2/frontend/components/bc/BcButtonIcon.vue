<script lang="ts" setup>
const props = defineProps<{
  /**
   *
   * ♿️ screenreader text every button with just an icon should describe what it does
   */
  srText: string,
}>()
</script>

<template>
  <button class="bc-button-icon">
    <span class="sr-only">{{ props.srText }}</span>
    <slot />
  </button>
</template>

<style lang="scss" scoped>
  .bc-button-icon {
    background-color: unset;
    border: unset;
}
</style>
