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
          :absolute="!absolute"
          :is-tooltip="true"
        />
      </slot>
    </template>
    <span v-if="absolute">
      <BcFormatNumber
        class="positive"
        :value="props.success"
      />
      <span class="slash"> / </span>
      <BcFormatNumber
        :class="data.failedClass"
        :value="props.failed"
      />
    </span>
    <BcFormatPercent
      v-else
      class="percent"
      :base="data.sum"
      :value="props.success"
      :fixed="undefined"
      :color-break-point="80"
      :full-on-empty-base="true"
    />
  </BcTooltip>
</template>
