<script setup lang="ts">
import type { BigNumber } from '@ethersproject/bignumber'
import type { ValueConvertOptions } from '~/types/value'

interface Props {
  value?: string | BigNumber
  options?: ValueConvertOptions
  useColors?: boolean
  positiveClass?: string
  negativeClass?: string,
  noTooltip?: boolean,
  fullValue?: boolean,
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
  const label = props.fullValue && res.fullLabel ? res.fullLabel : `${res.label}`
  if (props.useColors) {
    if (label.startsWith('-')) {
      labelClass = props.negativeClass
    } else if (res.label !== '0') {
      labelClass = props.positiveClass
    }
  }
  return {
    labelClass,
    label,
    fullLabel: res.fullLabel,
    tooltip: props.noTooltip ? '' : res.fullLabel
  }
})

</script>
<template>
  <BcTooltip>
    <template v-if="!!$slots.tooltip || data.tooltip" #tooltip>
      <slot name="tooltip" :data="data">
        <BcFormatNumber :text="data.tooltip" />
      </slot>
    </template>
    <span>
      <slot name="label" :data="data">
        <BcFormatNumber :class="data.labelClass" :text="data.label" />
      </slot>
    </span>
  </BcTooltip>
</template>
