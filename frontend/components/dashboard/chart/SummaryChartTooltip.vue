<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context
  startEpoch: number,
  groupInfos: {
    name: string,
    efficiency: number,
    color: string
  }[]
}

defineProps<Props>()

</script>

<template>
  <div class="tooltip-container">
    <DashboardChartTooltipHeader :t="t" :start-epoch="startEpoch" />
    <div v-for="(entry, index) in groupInfos" :key="index" class="line-container">
      <div class="circle" :style="{ 'background-color': entry.color }" />
      <div>
        {{ entry.name }}:
      </div>
      <div class="efficiency">
        {{ entry.efficiency }}%
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

  .line-container{
    display: flex;
    align-items: center;
    gap: 3px;

    .circle{
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    .efficiency{
      @include fonts.tooltip_text;
    }
  }
}
</style>
