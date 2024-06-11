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
import { isPublicKey } from '~/utils/dashboard/key'

const { isLoggedIn } = useUserStore()

const { dashboardKey, setDashboardKey } = useDashboardKeyProvider('validator')
const { refreshDashboards, updateHash, dashboards, cookieDashboards, getDashboardLabel } = useUserDashboardStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
// when we run into an error loading a dashboard keep it here to prevent an infinity loop
const errorDashboardKeys: string[] = []

const seoTitle = computed(() => {
  return getDashboardLabel(dashboardKey.value, 'validator')
})

useBcSeo(seoTitle, true)

const { refreshOverview, overview } = useValidatorDashboardOverviewStore()
await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [isLoggedIn] })

const { error: validatorOverviewError } = await useAsyncData('validator_overview', () => refreshOverview(dashboardKey.value), { watch: [dashboardKey] })
watch(validatorOverviewError, (error) => {
  if (error && dashboardKey.value) {
    if (!errorDashboardKeys.includes(dashboardKey.value)) {
      errorDashboardKeys.push(dashboardKey.value)
    }
    setDashboardKey('')
  }
}, { immediate: true })

const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog () {
  dashboardCreationControllerModal.value?.show()
}

const setDashboardKeyIfNoError = (key: string) => {
  if (!errorDashboardKeys.includes(key)) {
    setDashboardKey(key)
  }
}

watch([dashboardKey, isLoggedIn], ([newKey, newLoggedIn], [oldKey]) => {
  if (!newLoggedIn || !newKey) {
    // Some checks if we need to update the dashboard key or the public dashboard
    let cd = dashboards.value?.validator_dashboards?.[0] as CookieDashboard
    const isPublic = isPublicKey(newKey)
    if (newLoggedIn) {
      // if we are logged in and have no dashboard key we only want to switch to the first dashboard if it is a private one
      if (!cd.hash) {
        setDashboardKeyIfNoError(cd.id.toString())
      }
    } else if (!newLoggedIn && cd && isPublic && (!cd.hash || (cd.hash ?? '') === (oldKey ?? ''))) {
      // we got a new public dashboard hash but the old hash matches the stored dashboard - so we update the stored dashboard
      if (!errorDashboardKeys.includes(newKey)) {
        updateHash('validator', newKey)
      }
    } else if (!newKey || !isPublic) {
      // trying to view a private dashboad but not logged in
      cd = cookieDashboards.value?.validator_dashboards?.[0] as CookieDashboard
      setDashboardKeyIfNoError(cd?.hash ?? '')
    }
  }
}, { immediate: true })
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
      <DashboardSharedDashboardModal />
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
