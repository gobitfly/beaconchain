<script setup lang="ts">
import type { BigNumber } from '@ethersproject/bignumber'
import type { ValueConvertOptions } from '~/types/value'

interface Props {
  fullValue?: boolean,
  negativeClass?: string,
  noTooltip?: boolean,
  options?: ValueConvertOptions,
  positiveClass?: string,
  useColors?: boolean,
  value?: BigNumber | string,
}
const props = withDefaults(defineProps<Props>(), {
  negativeClass: 'negative',
  options: undefined,
  positiveClass: 'positive',
  value: undefined,
})

const { converter } = useValue()

const data = computed(() => {
  if (!props.value) {
    return {
      label: '',
      tooltip: '',
    }
  }
  const res = converter.value.weiToValue(props.value, props.options)
  let labelClass = ''
  const label
    = props.fullValue && res.fullLabel ? res.fullLabel : `${res.label}`
  if (props.useColors) {
    if (label.startsWith('-')) {
      labelClass = props.negativeClass
    }
    else if (res.label !== '0') {
      labelClass = props.positiveClass
    }
  }
  return {
    fullLabel: res.fullLabel,
    label,
    labelClass,
    tooltip: props.noTooltip ? '' : res.fullLabel,
  }
})
</script>

<template>
  <BcTooltip>
    <template
      v-if="!!$slots.tooltip || data.tooltip"
      #tooltip
    >
      <slot
        name="tooltip"
        :data
      >
        <BcFormatNumber :text="data.tooltip" />
      </slot>
    </template>
    <span>
      <BcFormatNumber
        :class="data.labelClass"
        :text="data.label"
      />
    </span>
  </BcTooltip>
</template>
