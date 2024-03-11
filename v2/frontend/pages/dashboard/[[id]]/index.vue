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
      <div class="panel-container">
        <DashboardCreationController ref="dashboardCreationControllerPanel" :display-type="'panel'" />
      </div>
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController ref="dashboardCreationControllerModal" :display-type="'modal'" />
    <BcPageWrapper>
      <template #top>
        <div class="header-container">
          <div class="h1 dashboard-title">
            {{ $t('dashboard.title') }}
          </div>
          <Button class="p-button-icon-only" @click="showDashboardCreation('modal')">
            <IconPlus alt="Plus icon" width="100%" height="100%" />
          </Button>
        </div>
        <DashboardValidatorOverview class="overview" :dashboard-key="key" />
      </template>
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

.header-container {
  display: flex;
  justify-content: space-between;

  .dashboard-title {
    margin-bottom: var(--padding-large);
  }
}

.panel-container {
  display: flex;
  justify-content: center;
  padding: 60px;
}

.overview {
  margin-bottom: var(--padding-large);
}

</style>
