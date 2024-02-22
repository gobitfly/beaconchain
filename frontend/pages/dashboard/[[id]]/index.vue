<script setup lang="ts">
import { warn } from 'vue'

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
    <TabView>
      <TabPanel header="Summary">
        Summary coming soon!
      </TabPanel>
      <TabPanel header="Rewards">
        Rewards coming soon!
      </TabPanel>
      <TabPanel header="Blocks">
        Blocks coming soon!
      </TabPanel>
      <TabPanel header="Heatmap">
        Heatmap coming soon!
      </TabPanel>
      <TabPanel header="Deposits">
        Deposits coming soon!
      </TabPanel>
      <TabPanel header="Withdrawals">
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
