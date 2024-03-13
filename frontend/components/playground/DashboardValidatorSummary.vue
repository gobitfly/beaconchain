<script setup lang="ts">
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'

const { getOverview } = useValidatorDashboardOverviewStore()
await useAsyncData('validator_dashboard_overview', () => getOverview(100))

</script>
<template>
  <DashboardTableSummary :dashboard-key="1" />
</template>

<style lang="scss" scoped>
</style>
