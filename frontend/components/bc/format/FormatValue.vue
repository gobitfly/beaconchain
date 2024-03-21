<script setup lang="ts">
import type { BigNumber } from '@ethersproject/bignumber'
import type { ValueConvertOptions } from '~/types/value'

interface Props {
  value?: string | BigNumber
  options?: ValueConvertOptions
  useColors?: boolean
  positiveClass?: string
  negativeClass?: string
}
const props = withDefaults(defineProps<Props>(), { value: undefined, options: undefined, positiveClass: 'positive', negativeClass: 'negative' })

const { converter } = useValue()

const data = computed(() => {
  if (!props.value) {
    return {
      label: '',
      tooltip: ''
    }
  }
  const res = converter.value.weiToValue(props.value, props.options)
  let labelClass = ''
  if (props.useColors) {
    if (`${res.label}`.startsWith('-')) {
      labelClass = props.negativeClass
    } else if (res.label !== '0') {
      labelClass = props.positiveClass
    }
  }
  return {
    labelClass,
    label: res.label,
    tooltip: res.fullLabel
  }
})

</script>
<template>
  <BcTooltip :text="data.tooltip">
    <template v-if="!!$slots.tooltip || data.tooltip" #tooltip>
      <slot name="tooltip" :data="data">
        {{ data.tooltip }}
      </slot>
    </template>
    <span :class="data.labelClass">
      {{ data.label }}
    </span>
  </BcTooltip>
</template>
