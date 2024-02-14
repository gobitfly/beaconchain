<script setup lang="ts">
interface Props {
  percent?: number
  base?: number
  value?: number
  hideEmptyValue?: boolean
  precision?: number
  fixed?: number
  colorBreakPoint?: number // if set then the percenage will be colored accordingly
}

const props = defineProps<Props>()

const data = computed(() => {
  let label: string | null = null
  let className = ''
  if (props.percent === undefined && !props.base) {
    if (!props.hideEmptyValue) {
      label = '0%'
    }
    return { label, className }
  }
  const percent = props.percent ?? calculatePercent(props.value, props.base)
  const config = { precision: props.precision ?? 2, fixed: props.fixed }
  label = formatPercent(percent, config)

  if (props.colorBreakPoint) {
    if ((props.base === 0 && percent === 0) || percent >= 80) {
      className = 'text-positive'
    } else {
      className = 'text-negative'
    }
  }
  return { label, className }
})

</script>
<template>
  <span :class="data.className">{{ data.label }}</span>
</template>
