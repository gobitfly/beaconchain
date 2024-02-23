<script setup lang="ts">
import { warn } from 'vue'
import {
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'

const route = useRoute()

warn(`route id: ${route.params.id}, query: ${route.query}`)
</script>

<template>
  <BcPageWrapper>
    <template #top>
      <div class="h1 dashboard_title">
        {{ $t('dashboard.title') }}
      </div>
      <DashboardValidatorOverview class="overview" />
    </template>
    <div>
      <DashboardValidatorSlotViz />
    </div>
    <TabView lazy>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.validator.tabs.summary')" :icon="faChartLineUp" />
        </template>
        <DashboardTableSummary :dashboard-id="1" />
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

.content {
  margin-bottom: var(--padding-large);
}

.dashboard_title, .overview{
  margin-bottom: var(--padding-large);
}

</style>
