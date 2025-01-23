<script setup lang="ts">
import type { BigNumber } from '@ethersproject/bignumber'
import type { ValueConvertOptions } from '~/types/value'

interface Props {
  fullValue?: boolean,
  hasColor?: boolean,
  noTooltip?: boolean,
  options?: ValueConvertOptions,
  value?: BigNumber | string,
}
const props = defineProps<Props>()

// const data = computed(() => {
// if (!props.value) {
//   return {
//     label: '',
//     tooltip: '',
//   }
// }
// const labelClass = ''
// const label
//   = props.fullValue && res.fullLabel ? res.fullLabel : `${res.label}`
// if (props.useColors) {
//   // else if (res.label !== '0') {
//   //   labelClass = props.positiveClass
//   // }
// }
// return {
//   fullLabel: res.fullLabel,
//   label,
//   labelClass,
//   tooltip: props.noTooltip ? '' : res.fullLabel,
// }
// })
const data = {
  label: 'xxxxx',
  tooltip: 'yyyy',
}
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
        :class="{
          positive: hasColor && data.label.startsWith('+'),
          negative: hasColor && data.label.startsWith('-'),
        }"
        :text="data.label"
      />
    </span>
  </BcTooltip>
</template>
