<script setup lang="ts">
import type { BigNumber } from '@ethersproject/bignumber'
import type { ValueConvertOptions } from '~/types/value'

interface Props {
  value?: string | BigNumber
  options? : ValueConvertOptions
}
const props = defineProps<Props>()

const { converter } = useValue()

const data = computed(() => {
  if (!props.value) {
    return {
      label: '',
      tooltip: ''
    }
  }
  const res = converter.value.weiToValue(props.value, props.options)
  return {
    label: res.label,
    tooltip: res.fullLabel
  }
})

</script>
<template>
  <BcTooltip :text="data.tooltip">
    {{ data.label }}
  </BcTooltip>
</template>
