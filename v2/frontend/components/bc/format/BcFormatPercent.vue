<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowDown,
  faArrowsLeftRight,
  faArrowUp,
} from '@fortawesome/pro-solid-svg-icons'
import type { CompareResult } from '~/types/value'

const {
  addPositiveSign,
  base,
  // if set then the percentage will be colored accordingly. Do not use it in combination with comparePercent
  colorBreakPoint,
  comparePercent,
  fullOnEmptyBase,
  hideEmptyValue,
  maximumFractionDigits = 2,
  minimumFractionDigits = 2,
  percent,
  trailingZeroDisplay = 'stripIfInteger',
  value,
}
 = defineProps<{
   addPositiveSign?: boolean,
   base?: number,
   // if set then the percentage will be colored accordingly. Do not use it in combination with comparePercent
   colorBreakPoint?: number,
   comparePercent?: number, // if set it adds the compare sign in front and colors the values accordingly
   fullOnEmptyBase?: boolean,
   hideEmptyValue?: boolean,
   maximumFractionDigits?: Intl.NumberFormatOptions['maximumFractionDigits'],
   minimumFractionDigits?: Intl.NumberFormatOptions['minimumFractionDigits'],
   percent?: number,
   trailingZeroDisplay?: Intl.NumberFormatOptions['trailingZeroDisplay'],
   value?: number,
 }>()

const data = computed(() => {
  let label: null | string = null
  let compareResult: CompareResult | null = null
  let className = ''
  if (base === 0 && fullOnEmptyBase) {
    return {
      className: 'text-positive',
      label: '100%',
    }
  }
  let leadingIcon: IconDefinition | undefined
  if (percent === undefined && !base) {
    if (!hideEmptyValue) {
      label = '0%'
    }
    return {
      className,
      label,
    }
  }
  const localPercent = percent ?? calculatePercent(value, base)
  label = new Intl.NumberFormat('en', {
    maximumFractionDigits,
    minimumFractionDigits,
    style: 'unit',
    trailingZeroDisplay,
    unit: 'percent',
  }).format(localPercent)
  label = addPositiveSign ? `+${label}` : label

  if (comparePercent !== undefined) {
    const thresholdToDifferenciateUnderperformerAndOverperformer = 0.25
    if (Math.abs(comparePercent - localPercent) <= thresholdToDifferenciateUnderperformerAndOverperformer) {
      className = 'text-equal'
      leadingIcon = faArrowsLeftRight
      compareResult = 'equal'
    }
    else if (localPercent > comparePercent) {
      className = 'text-positive'
      leadingIcon = faArrowUp
      compareResult = 'higher'
    }
    else {
      className = 'text-negative'
      leadingIcon = faArrowDown
      compareResult = 'lower'
    }
  }
  else if (colorBreakPoint) {
    if (
      (base === 0 && localPercent === 0)
      || localPercent >= colorBreakPoint
    ) {
      className = 'text-positive'
    }
    else {
      className = 'text-negative'
    }
  }
  return {
    className,
    compareResult,
    label,
    leadingIcon,
  }
})
</script>

<template>
  <span
    :class="data.className"
    class="format-percent"
  >
    <BcTooltip
      v-if="data.leadingIcon"
      class="direction-icon"
    >
      <template #tooltip>
        <slot
          name="leading-tooltip"
          v-bind="{ compare: data.compareResult }"
        />
      </template>
      <FontAwesomeIcon :icon="data.leadingIcon" />
    </BcTooltip>
    <BcFormatNumber
      v-if="data.label"
      :text="data.label"
    />
  </span>
</template>

<style lang="scss" scoped>
.format-percent {
  &:has(.direction-icon) {
    display: inline-flex;
    align-items: center;
    gap: 7px;
  }

  .direction-icon {
    display: inline-flex;
    justify-content: center;

    svg {
      height: 14px;
      width: auto;
    }
  }
}
</style>
