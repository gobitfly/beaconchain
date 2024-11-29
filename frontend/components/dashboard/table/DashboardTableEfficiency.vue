<script setup lang="ts">
import BcTooltip from '~/components/bc/BcTooltip.vue'

interface Props {
  absolute?: boolean,
  failed: number,
  isTooltip?: boolean,
  success: number,
}
const props = defineProps<Props>()

const data = computed(() => {
  const failedClass = props.failed ? 'negative' : 'positive'
  const sum = props.failed + props.success

  return {
    failedClass,
    sum,
  }
})
</script>

<template>
  <BcTooltip
    class="efficiency"
    :fit-content="true"
  >
    <template
      v-if="!isTooltip"
      #tooltip
    >
      <slot name="tooltip">
        <DashboardTableEfficiency
          v-bind="props"
          :absolute
          :is-tooltip="true"
        />
      </slot>
    </template>
    <span v-if="absolute">
      <span v-if="props.success === 0 && props.failed === 0">0 / 0</span>
      <template v-else>
        <BcFormatNumber
          class="positive"
          :value="props.success === 0 ? undefined : props.success"
        />
        <span class="slash"> / </span>
        <BcFormatNumber
          :class="data.failedClass"
          :value="props.failed === 0 ? undefined : props.failed"
        />
      </template>
    </span>
    <span v-else>
      <span v-if="props.success === 0 && props.failed === 0">-</span>
      <BcFormatPercent
        v-else
        class="percent"
        :base="typeof data.sum === 'number' ? data.sum : undefined"
        :value="props.success === 0 ? undefined : props.success"
        :fixed="undefined"
        :color-break-point="80"
        :full-on-empty-base="true"
      />
    </span>
  </BcTooltip>
</template>
