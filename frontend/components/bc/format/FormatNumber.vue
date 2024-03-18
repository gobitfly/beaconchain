<script setup lang="ts">

interface Props {
  value?: number | string,
  minDecimals?: number, // defaults to 0
  maxDecimals?: number, // defaults to 2
  default?: number | string // used if value is not defined
}
const props = defineProps<Props>()

const label = computed(() => {
  if (props.value === undefined || props.value === '') {
    return props.default
  }
  return trim(props.value, props.maxDecimals ?? 2, props.minDecimals ?? 0)
})

</script>
<template>
  <span v-if="label"> {{ label }}
  </span>
</template>
