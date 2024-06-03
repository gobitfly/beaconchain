<script setup lang="ts">
import {
  faArrowDown,
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { DashboardCreationController } from '#components'
import type { CookieDashboard } from '~/types/dashboard'

const { isLoggedIn } = useUserStore()

const { dashboardKey, setDashboardKey } = useDashboardKeyProvider('validator')
const { refreshDashboards, updateHash, dashboards, getDashboardLabel } = useUserDashboardStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const seoTitle = computed(() => {
  return getDashboardLabel(dashboardKey.value, 'validator')
})

useBcSeo(seoTitle, true)

const { refreshOverview, overview } = useValidatorDashboardOverviewStore()
await Promise.all([
  useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [isLoggedIn] }),
  useAsyncData('validator_overview', () => refreshOverview(dashboardKey.value), { watch: [dashboardKey] })
])

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
    // If the old key does not match the dashboards key then it probabbly means we opened a different public dashboard as a link
    if (cd && (!cd.hash || (cd.hash ?? '') === (oldKey ?? ''))) {
      updateHash('validator', newKey)
    }
  }
})
</script>

<template>
  <div v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <BcPageWrapper>
      <DashboardCreationController class="panel-controller" :display-type="'panel'" :initially-visislbe="true" />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-type="'modal'"
    />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader :dashboard-title="overview?.name" @show-creation="showDashboardCreationDialog()" />
        <DashboardValidatorOverview class="overview" />
      </template>
      <DashboardControls />
      <div>
        <DashboardValidatorSlotViz />
      </div>
      <TabView lazy class="dashboard-tab-view">
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
          <DashboardTableBlocks />
        </TabPanel>
        <TabPanel :disabled="!showInDevelopment">
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.heatmap')" :icon="faFire" />
          </template>
          Heatmap coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.deposits')" :icon="faWallet" />
          </template>
          <div class="deposits">
            <DashboardTableElDeposits />
            <FontAwesomeIcon :icon="faArrowDown" class="down_icon" />
            <DashboardTableClDeposits />
          </div>
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.withdrawals')" :icon="faMoneyBill" />
          </template>
          <DashboardTableWithdrawals />
        </TabPanel>
      </TabView>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>
.panel-controller {
  display: flex;
  justify-content: center;
  margin-top: 136px;
  margin-bottom: 307px;
  overflow: hidden;
}

:global(.dashboard-tab-view >.p-tabview-panels) {
  min-height: 699px;
}

:global(.modal-controller) {
  max-width: 100%;
  width: 460px;
}

.overview {
  margin-bottom: var(--padding-large);
}

.p-tabview {
  margin-top: var(--padding-large);
}

.down_icon {
  width: 100%;
  height: 28px;
}
</style>
