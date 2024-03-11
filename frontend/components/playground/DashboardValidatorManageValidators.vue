<script setup lang="ts">
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

const { getOverview } = useValidatorDashboardOverviewStore()
await useAsyncData('validator_dashboard_overview', () => getOverview())

const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards())

const selectedGroupId = ref<number>(DAHSHBOARDS_ALL_GROUPS_ID)

</script>
<template>
  <div class="icon_holder">
    <DashboardGroupSelection v-model="selectedGroupId" class="group_selection" />
    <DashboardGroupSelection v-model="selectedGroupId" class="group_selection" :include-all="true" />
    <div>{{ dashboards }}</div>
  </div>
</template>

<style lang="scss" scoped>
.group_selection{
  width: 200px;
}
.icon_holder {
  margin: 10px;
  display: flex;
  flex-direction: column;
  gap: var(--padding);
}
</style>
