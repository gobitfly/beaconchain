<script setup lang="ts">
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

const { getOverview } = useValidatorDashboardOverviewStore()
await useAsyncData('validator_dashboard_overview', () => getOverview(100))

const store = useUserDashboardStore()
const { getDashboards } = store

await useAsyncData('validator_dashboards', () => getDashboards())

const selectedGroupId = ref<number>(DAHSHBOARDS_ALL_GROUPS_ID)

</script>
<template>
  <div class="icon-holder">
    <DashboardGroupSelection v-model="selectedGroupId" class="group-selection" />
    <DashboardGroupSelection v-model="selectedGroupId" class="group-selection" :include-all="true" />
  </div>
  <div class="status-holder">
    <div class="status">
      <ValidatorTableStatus status="online" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="deposited" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="offline" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="pending" :position="12345" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="exited" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="slashed" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.group-selection{
  width: 200px;
}
.icon-holder {
  margin: 10px;
  display: flex;
  flex-direction: column;
  gap: var(--padding);
}

.status-holder{
  display: flex;
  flex-wrap: wrap;
  padding: 10px;
  .status{
    width: 140px;
    padding: 5px;
    border: 1px solid black;
  }
}
</style>
