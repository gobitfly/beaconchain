<script setup lang="ts">
export type Rotation = '0deg' | '90deg'

const {
  rotation = '0deg',
} = defineProps<{
  rotation?: Rotation,
}>()
</script>

<template>
  <span
    class="bc-rotate"
  >
    <slot />
  </span>
</template>

<style lang="scss" scoped>
.bc-rotate {
  transition: transform 0.2s;
  transform: rotate(v-bind(rotation));
}
</style>
