<script setup lang="ts">
interface Props {
  default?: number | string, // used if value is not defined
  maxDecimals?: number, // defaults to 2
  minDecimals?: number, // defaults to 0
  text?: string, // for already formatted numbers
  value?: number | string, // can either be a number or a string representing a number
}
const props = defineProps<Props>()

const label = computed(() => {
  if (props.text?.length) {
    return formattedNumberToHtml(props.text)
  }
  if (props.value === undefined || props.value === '') {
    return props.default
  }
  return formattedNumberToHtml(
    trim(props.value, props.maxDecimals ?? 2, props.minDecimals ?? 0),
  )
})
</script>

<template>
  <!-- eslint-disable vue/no-v-html -->
  <span
    v-if="label"
    v-html="label"
  />
  <!-- eslint-enable vue/no-v-html -->
</template>
