<script lang="ts" setup>
import SummaryChart from '../dashboard/chart/SummaryChart.vue'
</script>

<template>
  <div class="chart-container">
    <SummaryChart :dashboard-id="1" />
  </div>
</template>

<style lang="scss">
  .chart-container {
    width: 100%;
    height: 625px;
    padding: 0;
  }
</style>
