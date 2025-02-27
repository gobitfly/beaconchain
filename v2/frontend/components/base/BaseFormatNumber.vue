<script setup lang="ts">
import type { Color } from '~/components/base/BaseFormat.vue'

defineProps<{
  color?: Color,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  value: number | string,
}>()
</script>

<template>
  <BaseFormat :color>
    {{ formatNumber(value, {
      maximumFractionDigits,
      minimumFractionDigits,
    }) }}
  </BaseFormat>
</template>

<style scoped lang="scss">
</style>
