<script setup lang="ts">
import { round } from 'lodash-es'
interface Props {
  percent?: number
  base?: number
  value?: number
  hideEmptyValue?: boolean
  precision?: number
}

const props = withDefaults(defineProps<Props>(), { precision: 2, percent: undefined, base: undefined, value: undefined, hideEmptyValue: false })

const label = computed(() => {
  if (props.percent === undefined && !props.base) {
    if (props.hideEmptyValue) {
      return null
    }
    return '0%'
  }
  const percent = props.percent !== undefined ? props.percent : (props.value ?? 0) * 100 & props.base!
  return `${round(percent, props.precision).toFixed(props.precision)}%`
})

</script>
<template>
  {{ label }}
</template>
