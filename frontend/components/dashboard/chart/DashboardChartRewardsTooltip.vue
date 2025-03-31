<script lang="ts" setup>
defineProps<{
  consensusLayerRewardSum: string,
  consesnsusLayerRewardSumLabel: string,
  currentEpoch: {
    index: string,
    timestamp: number,
  },
  executionLayerRewardSum: string,
  executionLayerRewardSumLabel: string,
  groupInfo: {
    cl: { id: number, name: string, value: string }[],
    el: { id: number, name: string, value: string }[],
  },
}>()
</script>

<template>
  <div class="tooltip-container">
    <div>
      {{ getDateTime(currentEpoch.timestamp) }}
    </div>
    <div>
      Epoch: {{ currentEpoch.index }}
    </div>
    <div>
      <div class="header">
        <span
          class="circle cl"
        /><b>{{ consesnsusLayerRewardSumLabel }}: {{ consensusLayerRewardSum }}</b>
      </div>
      <ol>
        <li
          v-for="group in groupInfo.cl"
          :key="group.id"
        >
          {{ group.name }}: {{ group.value }}
        </li>
      </ol>
    </div>
    <div>
      <div class="header">
        <span
          class="circle el"
        /><b>{{ executionLayerRewardSumLabel }}: {{ executionLayerRewardSum }}</b>
      </div>
      <ol>
        <li
          v-for="group in groupInfo.el"
          :key="group.id"
        >
          {{ group.name }}: {{ group.value }}
        </li>
      </ol>
    </div>
  </div>
</template>

<style lang="scss">
@use "~/assets/css/fonts.scss";

.tooltip-container {
  @include fonts.tooltip_text_bold;
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  line-height: 1.5;
  padding: var(--padding);
  max-height: 400px;
  overflow-y: auto;
  pointer-events: all;

  .header {
    display: flex;
    align-items: center;
    margin-top: var(--padding);
    gap: 3px;

    .circle {
      width: 9px;
      height: 9px;
      border-radius: 50%;
      margin-bottom: 2px;

      &.el {
        background-color: var(--primary-orange);
      }

      &.cl {
        background-color: var(--melllow-blue);
      }
    }
  }

  ol {
    margin-block-start: 0;
    margin-block-end: 0;
    margin-inline-start: 0px;
    margin-inline-end: 0px;
    padding-inline-start: 26px;
  }
}
</style>
