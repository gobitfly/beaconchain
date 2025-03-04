<script setup lang="ts">
export type Color = 'gray' | 'green' | 'red'
defineProps<{
  color?: Color,
}>()
</script>

<template>
  <span
    :class="{
      green: color === 'green',
      red: color === 'red',
      gray: color === 'gray',
    }"
  >
    <slot />
  </span>
</template>

<style scoped lang="scss">
  .green {
    color: var(--positive-color);
  }
  .red {
    color: var(--negative-color);
  }
  .gray {
    color: var(--grey);
  }
</style>
