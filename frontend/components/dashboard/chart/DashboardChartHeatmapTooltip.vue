<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import type { WeiToValue } from '~/types/value'
import type { VDBHeatmapTooltipData } from '~/types/api/validator_dashboard'
import type { StatusCount } from '~/types/api/common'

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
    attestationIncome: `${props.weiToValue(props.tooltipData?.attestation_income)?.label}`,
    hasAttestation: props.tooltipData?.attestation_income || props.tooltipData?.attestation_efficiency,
    hasAction: props.tooltipData?.proposers || props.tooltipData?.slashings || props.tooltipData?.syncs
  }
})

</script>

<template>
  <div class="tooltip-container">
    <div :class="{ 'has-action': mapped.hasAction }" class="wrapper">
      <DashboardChartTooltipHeader :t="t" :start-epoch="startEpoch" />
      <div v-if="props.tooltipData?.proposers" class="row top">
        <div class="circle" :style="{ backgroundColor: colors.proposal }" />
        <span>{{ t('dashboard.validator.heatmap.proposers') }}:</span>
        <BcTableStatusCount class="value" :count="(props.tooltipData?.proposers as any as StatusCount)" />
      </div>
      <div v-if="props.tooltipData?.slashings" class="row top">
        <div class="circle" :style="{ backgroundColor: colors.slashing }" />
        <span>{{ t('dashboard.validator.heatmap.slashings') }}:</span>
        <BcTableStatusCount class="value" :count="(props.tooltipData?.slashings as any as StatusCount)" />
      </div>
      <div v-if="props.tooltipData?.syncs" class="row top">
        <div class="circle" :style="{ backgroundColor: colors.sync }" />
        <span>{{ t('dashboard.validator.heatmap.syncs') }}:</span>
        <BcFormatNumber class="value" :value="(props.tooltipData?.syncs as any as number)" />
      </div>
      <div v-if="mapped.hasAttestation">
        <div v-if="mapped.attestationIncome !== undefined" class="row">
          <span>{{ t('dashboard.validator.heatmap.attestations_income') }}:</span>
          <BcFormatNumber class="value" :text="mapped.attestationIncome" />
        </div>
        <div v-if="props.tooltipData?.attestation_efficiency !== undefined" class="row">
          <span>{{ t('dashboard.validator.heatmap.attestation_efficiency') }}:</span>
          <BcFormatPercent class="value" :percent="props.tooltipData?.attestation_efficiency" />
        </div>
      </div>
      <div v-if="mapped.hasAttestation">
        <div v-if="props.tooltipData?.attestations_head !== undefined" class="row">
          <BcTableStatusCount class="value" :count="props.tooltipData?.attestations_head" />
          <span>{{ t('dashboard.validator.heatmap.attestations_head') }}</span>
        </div>
        <div v-if="props.tooltipData?.attestations_source !== undefined" class="row">
          <BcTableStatusCount class="value" :count="props.tooltipData?.attestations_source" />
          <span>{{ t('dashboard.validator.heatmap.attestations_source') }}</span>
        </div>
        <div v-if="props.tooltipData?.attestations_target !== undefined" class="row">
          <BcTableStatusCount class="value" :count="props.tooltipData?.attestations_target" />
          <span>{{ t('dashboard.validator.heatmap.attestations_target') }}</span>
        </div>
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

  .wrapper {
    &.has-action {
      margin-left: 19px;
    }

    >div {
      margin-top: var(--padding-small);
    }

    .row {
      position: relative;
      display: flex;
      gap: var(--padding-small);

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
}
</style>
