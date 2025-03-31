<script setup lang="ts">
export type Color =
  | 'gray'
  | 'green'
  | 'red'

defineProps<{
  color?: Color,
}>()
</script>

<template>
  <span
    :class="{
      'bc-color--gray': color === 'gray',
      'bc-color--green': color === 'green',
      'bc-color--red': color === 'red',
    }"
  >
    <slot />
  </span>
</template>

<style scoped lang="scss">
  .bc-color--green {
    color: var(--positive-color);
  }
  .bc-color--red {
    color: var(--negative-color);
  }
  .bc-color--gray {
    color: var(--grey);
  }
</style>
