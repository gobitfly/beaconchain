<script setup lang="ts">
import { DashboardGroupSelectionDialog } from '#components'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

const overviewStore = useValidatorDashboardOverviewStore()
const { getOverview } = overviewStore
await useAsyncData('validator_dashboard_overview', () => getOverview(100))
const { overview } = storeToRefs(overviewStore)

const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards())

const selectedGroupId = ref<number>(DAHSHBOARDS_ALL_GROUPS_ID)

const dialog = useDialog()

function onClose (groupId: boolean) {
  setTimeout(() => {
    alert('new group: ' + groupId)
  }, 100
  )
}

const openGroupSelection = (withPreselection: boolean) => {
  dialog.open(DashboardGroupSelectionDialog, {
    onClose: response => onClose(response?.data),
    data: {
      groupId: withPreselection ? overview.value?.groups?.[0]?.id : undefined,
      selectedValidators: 10,
      totalValidators: 123
    }
  })
}

</script>
<template>
  <div class="icon_holder">
    <DashboardGroupSelection v-model="selectedGroupId" class="group_selection" />
    <DashboardGroupSelection v-model="selectedGroupId" class="group_selection" :include-all="true" />
    <div>{{ dashboards }}</div>
  </div>
  <div class="icon_holder">
    <Button class="group_selection" label="Open Group Selection preselected" @click="openGroupSelection(true)" />
    <Button class="group_selection" label="Open Group Selection" @click="openGroupSelection(false)" />
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
