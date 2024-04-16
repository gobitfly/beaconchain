<script setup lang="ts">
import {
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'
import type { DashboardCreationController } from '#components'
import { type CookieDashboard } from '~/types/dashboard'

const { dashboardKey, setDashboardKey, isPublic } = useDashboardKeyProvider()

const { isLoggedIn } = useUserStore()
const { refreshDashboards, updateHash, dashboards } = useUserDashboardStore()
const { refreshOverview } = useValidatorDashboardOverviewStore()
await Promise.all([
  useAsyncData('user_dashboards', () => refreshDashboards()),
  useAsyncData('validator_overview', () => refreshOverview(dashboardKey.value), { watch: [dashboardKey] })
])

const manageValidatorsModalVisisble = ref(false)
const manageGroupsModalVisisble = ref(false)

const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog () {
  dashboardCreationControllerModal.value?.show()
}

onMounted(() => {
  if (dashboardKey.value === '') {
    // we don't have a key and no validator dashboard: show the create panel
    if (dashboards.value?.validator_dashboards?.length) {
      // if we have a validator dashboard but none selected: select the first
      const cd = dashboards.value.validator_dashboards[0] as CookieDashboard
      setDashboardKey(cd.hash ?? cd.id.toString())
    }
  }
})

watch(dashboardKey, (newKey, oldKey) => {
  if (!isLoggedIn.value) {
    // We update the key for our public dashboard
    const cd = dashboards.value?.validator_dashboards?.[0] as CookieDashboard
    // If the old key does not match the dashboards key then it probabbly means we opened a different pub. dashboard as a link
    if (cd && (!cd.hash || (cd.hash ?? '') === (oldKey ?? ''))) {
      updateHash('validator', newKey)
    }
  }
})
</script>

<template>
  <div v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <BcPageWrapper>
      <DashboardCreationController
        class="panel-controller"
        :display-type="'panel'"
        :initially-visislbe="true"
      />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" />
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" />
    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-type="'modal'"
    />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardValidatorOverview class="overview" />
      </template>
      <div class="edit-buttons-row">
        <Button v-if="isLoggedIn && !isPublic" :label="$t('dashboard.validator.manage-groups')" @click="manageGroupsModalVisisble = true" />
        <Button :label="$t('dashboard.validator.manage-validators')" @click="manageValidatorsModalVisisble = true" />
      </div>
      <div>
        <DashboardValidatorSlotViz />
      </div>
      <TabView lazy>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.summary')" :icon="faChartLineUp" />
          </template>
          <DashboardTableSummary />
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.rewards')" :icon="faCubes" />
          </template>
          <DashboardTableRewards />
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.blocks')" :icon="faCube" />
          </template>
          <DashboardTableBlocks :dashboard-key="key" />
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.heatmap')" :icon="faFire" />
          </template>
          Heatmap coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.deposits')" :icon="faWallet" />
          </template>
          Deposits coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.withdrawals')" :icon="faMoneyBill" />
          </template>
          Withdrawals coming soon!
        </TabPanel>
      </TabView>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>
.edit-buttons-row{
  display: flex;
  justify-content: flex-end;
  gap: var(--padding);
  margin-bottom: var(--padding);
}
.panel-controller {
  display: flex;
  justify-content: center;
  margin-top: 60px;
  margin-bottom: 60px;
  overflow: hidden;
}

:global(.modal-controller) {
  max-width: 100%;
  width: 460px;
}

.overview {
  margin-bottom: var(--padding-large);
}
</style>
