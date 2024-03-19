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
import { type DashboardCreationDisplayType } from '~/types/dashboard/creation'

const route = useRoute()

const key = computed(() => {
  if (Array.isArray(route.params.id)) {
    return route.params.id.join(',')
  }
  return route.params.id
})

const manageValidatorsModalVisisble = ref(false)

const dashboardCreationControllerPanel = ref<typeof DashboardCreationController>()
const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreation (type: DashboardCreationDisplayType) {
  if (type === 'panel') {
    dashboardCreationControllerPanel.value?.show()
  } else {
    dashboardCreationControllerModal.value?.show()
  }
}

onMounted(() => {
  // TODO: Implement check if user does not have a single dashboard instead of the key check once information is available
  if (key.value === '') {
    showDashboardCreation('panel')
  }
})
</script>

<template>
  <div v-if="key==''">
    <BcPageWrapper>
      <DashboardCreationController ref="dashboardCreationControllerPanel" class="panel-controller" :display-type="'panel'" />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" :dashboard-key="key" />
    <DashboardCreationController ref="dashboardCreationControllerModal" class="modal-controller" :display-type="'modal'" />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader :dashboard-key="key" @show-creation="showDashboardCreation('modal')" />
        <DashboardValidatorOverview class="overview" :dashboard-key="key" />
      </template>
      <Button :label="$t('dashboard.validator.manage-validators')" @click="manageValidatorsModalVisisble = true" />
      <div>
        <DashboardValidatorSlotViz :dashboard-key="key" />
      </div>
      <TabView lazy>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.summary')" :icon="faChartLineUp" />
          </template>
          <DashboardTableSummary :dashboard-key="key" />
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.rewards')" :icon="faCubes" />
          </template>
          Rewards coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.blocks')" :icon="faCube" />
          </template>
          Blocks coming soon!
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

.panel-controller {
  display: flex;
  justify-content: center;
  padding: 60px 0px;
}

:global(.modal_controller) {
  max-width: 460px;
  width: 100%;
}

.overview {
  margin-bottom: var(--padding-large);
}

</style>
