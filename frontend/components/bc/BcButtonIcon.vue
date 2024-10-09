<script lang="ts" setup>
const props = defineProps<{
  /**
   *
   * ♿️ screenreader text
   * every button with just an icon has to describe what it does
   */
  screenreaderText: string,
}>()
</script>

<template>
  <button class="bc-button-icon">
    <BcScreenreaderOnly>{{ props.screenreaderText }}</BcScreenreaderOnly>
    <slot />
  </button>
</template>

<style lang="scss" scoped>
  .bc-button-icon {
    background-color: unset;
    border: unset;
}
</style>
