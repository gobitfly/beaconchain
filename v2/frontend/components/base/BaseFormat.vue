<script setup lang="ts">
import type { Color } from '~/components/bc/BcColor.vue'

defineProps<{
  color?: Color,
}>()
</script>

<template>
  <BcColor
    :color
  >
    <slot />
  </BcColor>
</template>

<style scoped lang="scss">
</style>
