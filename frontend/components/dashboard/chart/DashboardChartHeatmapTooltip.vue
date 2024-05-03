<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import type { WeiToValue } from '~/types/value'
import type { VDBHeatmapTooltipData, VDBHeatmapTooltipDuty } from '~/types/api/validator_dashboard'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context,
  weiToValue: WeiToValue,
  startEpoch: number,
  theme: string,
  tooltipData?: VDBHeatmapTooltipData
}

const props = defineProps<Props>()

const colors = getHeatmapContentColors(props.theme)

const mapDuties = (duties?: VDBHeatmapTooltipDuty[]) => {
  return duties?.length
    ? duties?.reduce((list, d) => {
      if (d.status === 'success') {
        list[0]++
      } else {
        list[1]++
      }
      return list
    }, [0, 0])
    : undefined
}

const mapped = computed(() => {
  return {
    attestationIncome: props.weiToValue(props.tooltipData?.attestation_income),
    proposers: mapDuties(props.tooltipData?.proposers),
    slashings: mapDuties(props.tooltipData?.slashings),
    syncs: props.tooltipData?.syncs?.length
      ? props.tooltipData?.syncs
      : undefined,
    hasAction: props.tooltipData?.proposers?.length || props.tooltipData?.slashings?.length || props.tooltipData?.syncs?.length
  }
})

</script>

<template>
  <div class="tooltip-container">
    <div :class="{ 'has-action': mapped.hasAction }">
      <DashboardChartTooltipHeader :t="t" :start-epoch="startEpoch" />
      <div v-if="mapped.proposers" class="row">
        <div class="circle" :style="{ backgroundColor: colors.proposal }" />
        <span>{{ t('dashboard.validator.heatmap.proposers') }}</span>
        <span class="value" :class="{ positive: !!mapped.proposers[0] }"><span>{{ mapped.proposers[0] }}</span>, <span
          :class="{ negative: !!mapped.proposers[1] }"
        >{{ mapped.proposers[1] ?? 0 }}</span></span>
      </div>
      <div v-if="mapped.slashings" class="row">
        <div class="circle" :style="{ backgroundColor: colors.slashing }" />
        <span>{{ t('dashboard.validator.heatmap.slashings') }}</span>
        <span class="value" :class="{ positive: !!mapped.slashings[0] }"><span>{{ mapped.slashings[0] }}</span>, <span
          :class="{ negative: !!mapped.slashings[1] }"
        >{{ mapped.slashings[1] ?? 0 }}</span></span>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
@use '~/assets/css/fonts.scss';

.tooltip-container {
  @include fonts.tooltip_text;
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  line-height: 1.5;
  padding: var(--padding);
  max-height: 400px;
  overflow-y: auto;
  pointer-events: all;

  .has-action {
    margin-left: 19px;
  }

  .row {
    position: relative;
    display: flex;
    gap: var(--padding);

    .circle {
      position: absolute;
      width: 14px;
      height: 14px;
      border-radius: 50%;
      left: -19px;
    }

    .value {
      font-weight: var(--tooltip_text_bold_font_weight);
    }
  }
}
</style>
