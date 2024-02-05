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
  const config = { precision: props.precision, fixed: props.fixed }
  if (props.percent !== undefined) {
    return formatPercent(props.percent, config)
  }
  return formatAndCalculatePercent(props.value, props.base, config)
})

</script>
<template>
  {{ label }}
</template>
