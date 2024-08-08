<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUp,
  faArrowDown,
  faArrowsLeftRight,
} from '@fortawesome/pro-solid-svg-icons'
import { type CompareResult } from '~/types/value'

interface Props {
  percent?: number
  base?: number
  value?: number
  comparePercent?: number // if set it adds the compare sign in front and colors the values accordingly
  hideEmptyValue?: boolean
  precision?: number
  fixed?: number
  fullOnEmptyBase?: boolean
  addPositiveSign?: boolean
  // if set then the percentage will be colored accordingly. Do not use it in combination with comparePercent
  colorBreakPoint?: number
}

const props = defineProps<Props>()

const data = computed(() => {
  let label: string | null = null
  let compareResult: CompareResult | null = null
  let className = ''
  if (props.base === 0 && props.fullOnEmptyBase) {
    return {
      label: '100%',
      className: 'text-positive',
    }
  }
  let leadingIcon: IconDefinition | undefined
  if (props.percent === undefined && !props.base) {
    if (!props.hideEmptyValue) {
      label = '0%'
    }
    return { label, className }
  }
  const percent = props.percent ?? calculatePercent(props.value, props.base)
  const config = {
    precision: props.precision ?? 2,
    fixed: props.fixed ?? 2,
    addPositiveSign: props.addPositiveSign,
  }
  label = formatPercent(percent, config)
  if (props.comparePercent !== undefined) {
    if (Math.abs(props.comparePercent - percent) <= 0.5) {
      className = 'text-equal'
      leadingIcon = faArrowsLeftRight
      compareResult = 'equal'
    }
    else if (percent > props.comparePercent) {
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
  else if (props.colorBreakPoint) {
    if (
      (props.base === 0 && percent === 0)
      || percent >= props.colorBreakPoint
    ) {
      className = 'text-positive'
    }
    else {
      className = 'text-negative'
    }
  }
  return { label, className, leadingIcon, compareResult }
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
