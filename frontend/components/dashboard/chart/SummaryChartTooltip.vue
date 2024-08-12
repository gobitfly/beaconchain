<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import {
  type AggregationTimeframe,
  type EfficiencyType,
} from '~/types/dashboard/summary'

interface Props {
  aggregation: AggregationTimeframe,
  efficiencyType: EfficiencyType,
  groupInfos: {
    color: string,
    efficiency: number,
    name: string,
  }[],
  highlightGroup?: string,
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context
  ts: number,
}

defineProps<Props>()
</script>

<template>
  <div class="tooltip-container">
    <DashboardChartTooltipHeader
      :t
      :ts
      :aggregation
      :efficiency-type
    />
    <div
      v-for="(entry, index) in groupInfos"
      :key="index"
      class="line-container"
      :class="{ highlight: entry.name === highlightGroup }"
    >
      <div
        class="circle"
        :style="{ 'background-color': entry.color }"
      />
      <div class="name">
        {{ entry.name }}:
      </div>
      <BcFormatPercent
        class="efficiency"
        :percent="entry.efficiency"
      />
    </div>
  </div>
</template>

<style lang="scss">
@use "~/assets/css/fonts.scss";

.tooltip-container {
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  border: 1px transparent solid;
  line-height: 1.5;
  padding: var(--padding);

  .line-container {
    display: flex;
    align-items: center;
    gap: 3px;

    .circle {
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    .name,
    .efficiency {
      @include fonts.tooltip_text;
    }

    &:not(.highlight) {
      .efficiency,
      .name {
        opacity: 0.5;
      }
    }

    &.highlight {
      .efficiency,
      .name {
        font-weight: var(--tooltip_text_bold_font_weight);
      }
    }
  }
}
</style>
