<script setup lang="ts">

interface Props {
  value?: number | string, // can either be a number or a string representing a number
  text?: string, // for already formatted numbers
  minDecimals?: number, // defaults to 0
  maxDecimals?: number, // defaults to 2
  default?: number | string, // used if value is not defined
}
const props = defineProps<Props>()

function renderer (props: Props) : Array<VNode|string> {
  if (props.text?.length) {
    return formattedNumberToVDOM(props.text)
  }
  if (props.value === undefined || props.value === '') {
    return props.default === undefined ? [] : [String(props.default)]
  }
  return formattedNumberToVDOM(trim(props.value, props.maxDecimals ?? 2, props.minDecimals ?? 0))
}
</script>

<template>
  <span><renderer v-bind="props" /></span>
</template>
