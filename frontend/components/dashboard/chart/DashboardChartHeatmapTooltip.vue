<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import type { WeiToValue } from '~/types/value'
import type { VDBHeatmapTooltipData } from '~/types/api/validator_dashboard'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context,
  weiToValue: WeiToValue,
  startEpoch: number,
  theme: string,
  tooltipData?: VDBHeatmapTooltipData
}

const props = defineProps<Props>()

const colors = getHeatmapContentColors(props.theme)

const mapped = computed(() => {
  return {
    attestationIncome: props.weiToValue(props.tooltipData?.attestation_income),
    proposers: props.tooltipData?.proposers?.length
      ? props.tooltipData?.proposers?.reduce((list, p) => {
        if (p.status === 'success') {
          list[0]++
        } else {
          list[1]++
        }
        return list
      }, [0, 0])
      : undefined,
    hasAction: props.tooltipData?.proposers?.length || props.tooltipData?.slashings?.length || props.tooltipData?.syncs?.length
  }
})

</script>

<template>
  <div class="tooltip-container">
    <div :class="{'has-action': mapped.hasAction}">
      <DashboardChartTooltipHeader :t="t" :start-epoch="startEpoch" />
      <div v-if="mapped.proposers" class="row">
        <div class="circle" :style="{backgroundColor: colors.proposal}" />{{ t('dashboard.validator.heatmap.proposers') }}
      </div>
    </div>
  </div>
</template>

<style lang="scss">
@use '~/assets/css/fonts.scss';

.tooltip-container {
  @include fonts.tooltip_text_bold;
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  line-height: 1.5;
  padding: var(--padding);
  max-height: 400px;
  overflow-y: auto;
  pointer-events: all;
  .has-action{
    margin-left: 19px;
  }

  .row{
    position: relative;
    .circle {
      position: relative;
      width: 14px;
      height: 14px;
      border-radius: 50%;
      left: -19px;
    }
  }
}
</style>
