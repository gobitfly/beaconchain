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

const result = computed(() => {
  return formatPercent(ratio.value, {
    maximumFractionDigits,
    minimumFractionDigits,
  })
})
</script>

<template>
  <span>
    <slot
      :result
      :ratio
    >
      <BcColor :color>
        {{ result }}
      </BcColor>
    </slot>
  </span>
</template>
