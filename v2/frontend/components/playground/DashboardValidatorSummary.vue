<script setup lang="ts">

const { getOverview } = useValidatorDashboardOverview()
await useAsyncData('validator_dashboard_overview', () => getOverview())

</script>
<template>
  <DashboardTableSummary :dashboard-id="1" />
</template>

<style lang="scss" scoped>
</style>
