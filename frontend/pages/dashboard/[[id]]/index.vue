<script setup lang="ts">
import {
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'

import { type DashboardCreationDisplayType } from '~/types/dashboard/creation'

const route = useRoute()

const key = computed(() => {
  if (Array.isArray(route.params.id)) {
    return route.params.id.join(',')
  }
  return route.params.id
})

const displayType = ref<DashboardCreationDisplayType>('') // TODO: Set to panel when no dashbaord is available

const onAddDashboard = () => {
  displayType.value = 'modal'
}

</script>

<template>
  <BcPageWrapper>
    <template #top>
      <div class="header-container">
        <div class="h1 dashboard-title">
          {{ $t('dashboard.title') }}
        </div>
        <Button class="p-button-icon-only" @click="onAddDashboard">
          <IconPlus alt="Plus icon" width="100%" height="100%" />
        </Button>
      </div>
      <DashboardValidatorOverview class="overview" />
    </template>
    <DashboardCreationController v-model="displayType" />
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
</template>

<style lang="scss" scoped>

.header-container {
  display: flex;
  justify-content: space-between;

  .dashboard-title {
    margin-bottom: var(--padding-large);
  }
}

.overview {
  margin-bottom: var(--padding-large);
}

.content {
  margin-bottom: var(--padding-large);
}

</style>
