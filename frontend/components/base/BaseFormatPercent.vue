<script setup lang="ts">
import type { Color } from '~/components/base/BaseFormat.vue'

const {
  base = 1,
  maximumFractionDigits = 2,
  minimumFractionDigits,
  value,
} = defineProps<{
  base?: number | string,
  color?: Color,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  value: number | string,
}>()

const ratio = computed(() => {
  assertIsNumber(value, base)
  assertIsNotZero(base)
  return (Number(value) / Number(base))
})
</script>

<template>
  <BaseFormat :color>
    <slot :value>
      {{ formatPercent(ratio, {
        maximumFractionDigits,
        minimumFractionDigits,
      }) }}
    </slot>
  </BaseFormat>
</template>

<style scoped lang="scss">
</style>
