<script setup lang="ts">
import {
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'

const route = useRoute()

const key = computed(() => {
  if (Array.isArray(route.params.id)) {
    return route.params.id.join(',')
  }
  return route.params.id
})

const manageValidatorsModalVisisble = ref(false)

</script>

<template>
  <BcPageWrapper>
    <template #top>
      <div class="h1 dashboard_title">
        {{ $t('dashboard.title') }}
      </div>
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
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" :dashboard-key="key" />
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
