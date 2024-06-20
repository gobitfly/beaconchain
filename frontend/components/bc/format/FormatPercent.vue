<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUp,
  faArrowDown,
  faArrowsLeftRight
} from '@fortawesome/pro-solid-svg-icons'

interface Props {
  percent?: number
  base?: number
  value?: number
  comparePercent?: number
  hideEmptyValue?: boolean
  precision?: number
  fixed?: number
  fullOnEmptyBase?: boolean
  addPositiveSign?: boolean
  colorBreakPoint?: number // if set then the percentage will be colored accordingly
}

const props = defineProps<Props>()

const data = computed(() => {
  let label: string | null = null
  let className = ''
  if (props.base === 0 && props.fullOnEmptyBase) {
    return {
      label: '100%',
      className: 'text-positive'
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
  const config = { precision: props.precision ?? 2, fixed: props.fixed ?? 2, addPositiveSign: props.addPositiveSign }
  label = formatPercent(percent, config)
  if (props.comparePercent !== undefined) {
    if (props.comparePercent.toFixed(1) === percent.toFixed(1)) {
      className = 'text-equal'
      leadingIcon = faArrowsLeftRight
    } else if (percent > props.comparePercent) {
      className = 'text-positive'
      leadingIcon = faArrowUp
    } else {
      className = 'text-negative'
      leadingIcon = faArrowDown
    }
  } else if (props.colorBreakPoint) {
    if ((props.base === 0 && percent === 0) || percent >= props.colorBreakPoint) {
      className = 'text-positive'
    } else {
      className = 'text-negative'
    }
  }
  return { label, className, leadingIcon }
})

</script>
<template>
  <span :class="data.className" class="format-percent">
    <span v-if="data.leadingIcon" class="direction-icon">
      <FontAwesomeIcon :icon="data.leadingIcon" />
    </span>
    <BcFormatNumber v-if="data.label" :text="data.label" />
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
    width: 13px;
    display: inline-flex;
    justify-content: center;

    svg {
      height: 8px;
      width: auto;
    }
  }
}
</style>
