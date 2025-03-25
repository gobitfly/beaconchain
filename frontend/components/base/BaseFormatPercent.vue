<script setup lang="ts">
import type { Color } from '~/components/bc/BcColor.vue'

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
  <BcColor :color>
    <slot :value>
      {{ formatPercent(ratio, {
        maximumFractionDigits,
        minimumFractionDigits,
      }) }}
    </slot>
  </BcColor>
</template>

<style scoped lang="scss">
</style>
