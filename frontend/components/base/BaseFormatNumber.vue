<script setup lang="ts">
import type { Color } from '~/components/bc/BcColor.vue'

defineProps<{
  color?: Color,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  value: number | string,
}>()
</script>

<template>
  <BcColor :color>
    {{ formatNumber(value, {
      maximumFractionDigits,
      minimumFractionDigits,
    }) }}
  </BcColor>
</template>

<style scoped lang="scss">
</style>
