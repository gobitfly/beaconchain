<template>
  <div class="chart-container">
    <DashboardChartSummaryChart :dashboard-key="5003" />
  </div>
  <div class="chart-container">
    <DashboardChartRewardsChart :dashboard-key="5003" />
  </div>
</template>

<style lang="scss" scoped>
  .chart-container {
    width: 100%;
    height: 625px;
    padding: 0;
  }
</style>
