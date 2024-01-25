<script setup lang="ts">
interface Props {
  percent?: number
  base?: number
  value?: number
  hideEmptyValue?: boolean
  precision?: number
  fixed?: number
}

const props = defineProps<Props>()

const label = computed(() => {
  if (props.percent === undefined && !props.base) {
    if (props.hideEmptyValue) {
      return null
    }
    return '0%'
  }
  return formatPercent({ percent: props.percent, base: props.base, value: props.value }, { precision: props.precision, fixed: props.fixed })
})

</script>
<template>
  {{ label }}
</template>
