<script setup lang="ts">
import {
  faArrowDown,
  faChartLineUp,
  faCube,
  faCubes, faFire,
  faMoneyBill,
  faWallet,
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  DashboardCreationController, DashboardTableBlocks, DashboardTableEmpty, DashboardTableRewards, DashboardTableSummary,
  DashboardTableWithdrawals,
} from '#components'
import type { CookieDashboard } from '~/types/dashboard'
import {
  isPublicDashboardKey, isSharedKey,
} from '~/utils/dashboard/key'
import type { HashTabs } from '~/types/hashTabs'

const { isLoggedIn } = useUserStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { t: $t } = useTranslation()

const tabs: HashTabs = [
  {
    component: DashboardTableSummary,
    icon: faChartLineUp,
    key: 'summary',
    title: $t('dashboard.validator.tabs.summary'),
  },
  {
    component: DashboardTableRewards,
    icon: faCubes,
    key: 'rewards',
    title: $t('dashboard.validator.tabs.rewards'),
  },
  {
    component: DashboardTableBlocks,
    icon: faCube,
    key: 'blocks',
    title: $t('dashboard.validator.tabs.blocks'),

  },
  {
    component: DashboardTableEmpty,
    disabled: !showInDevelopment,
    icon: faFire,
    key: 'heatmap',
    title: $t('dashboard.validator.tabs.heatmap'),
  },
  {
    icon: faWallet,
    key: 'deposits',
    title: $t('dashboard.validator.tabs.deposits'),
  },
  {
    component: DashboardTableWithdrawals,
    icon: faMoneyBill,
    key: 'withdrawals',
    title: $t('dashboard.validator.tabs.withdrawals'),
  },
]

const {
  dashboardKey, setDashboardKey,
} = useDashboardKeyProvider('validator')
const {
  cookieDashboards,
  dashboards,
  getDashboardLabel,
  refreshDashboards,
  updateHash,
} = useUserDashboardStore()
// when we run into an error loading a dashboard keep it here to prevent an infinity loop
const errorDashboardKeys: string[] = []

const seoTitle = computed(() => {
  return getDashboardLabel(dashboardKey.value, 'validator')
})

useBcSeo(seoTitle, true)

const {
  overview, refreshOverview,
} = useValidatorDashboardOverviewStore()
await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

const { error: validatorOverviewError } = await useAsyncData(
  'validator_overview',
  () => refreshOverview(dashboardKey.value),
  { watch: [ dashboardKey ] },
)
watch(
  validatorOverviewError,
  (error) => {
    // we temporary blacklist dashboard id's that threw an error
    if (
      error
      && dashboardKey.value
      && !(
        !!dashboards.value?.account_dashboards?.find(
          d => d.id.toString() === dashboardKey.value,
        )
        || !!dashboards.value?.validator_dashboards?.find(
          d => !d.is_archived && d.id.toString() === dashboardKey.value,
        )
      )
    ) {
      if (!errorDashboardKeys.includes(dashboardKey.value)) {
        errorDashboardKeys.push(dashboardKey.value)
      }
      setDashboardKey('')
    }
  },
  { immediate: true },
)

const dashboardCreationControllerModal
  = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog() {
  dashboardCreationControllerModal.value?.show()
}

const setDashboardKeyIfNoError = (key: string) => {
  if (!errorDashboardKeys.includes(key)) {
    setDashboardKey(key)
  }
}

watch(
  [
    dashboardKey,
    isLoggedIn,
  ],
  ([
    newKey,
    newLoggedIn,
  ], [ oldKey ]) => {
    if (!newLoggedIn || !newKey) {
      // Some checks if we need to update the dashboard key or the public dashboard
      let cd = dashboards.value?.validator_dashboards?.[0] as CookieDashboard
      const isPublic = isPublicDashboardKey(newKey)
      const isShared = isSharedKey(newKey)
      if (isShared) {
        return
      }
      if (newLoggedIn) {
        // if we are logged in and have no dashboard key we only want to switch
        //  to the first dashboard if it is a private one
        if (cd && cd.hash === undefined) {
          setDashboardKeyIfNoError(cd.id.toString())
        }
      }
      else if (
        !newLoggedIn
        && cd
        && isPublic
        && (!cd.hash || (cd.hash ?? '') === (oldKey ?? ''))
      ) {
        // we got a new public dashboard hash but the old hash matches the
        // stored dashboard - so we update the stored dashboard
        if (!errorDashboardKeys.includes(newKey)) {
          updateHash('validator', newKey)
        }
        setDashboardKeyIfNoError(newKey ?? '')
      }
      else if (!newKey || !isPublic) {
        // trying to view a private dashboad but not logged in
        cd = cookieDashboards.value
          ?.validator_dashboards?.[0] as CookieDashboard
        setDashboardKeyIfNoError(cd?.hash ?? '')
      }
    }
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <BcPageWrapper>
      <DashboardCreationController class="panel-controller" :display-mode="'panel'" :initially-visible="true" />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController
      ref="dashboardCreationControllerModal" class="modal-controller"
      :display-mode="'modal'"
    />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardControls :dashboard-title="overview?.name" />
        <DashboardValidatorOverview class="overview" />
      </template>
      <DashboardSharedDashboardModal />
      <div>
        <DashboardValidatorSlotViz />
      </div>
      <BcTabList
        :tabs default-tab="summary"
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
      >
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits />
            <FontAwesomeIcon :icon="faArrowDown" class="down_icon" />
            <DashboardTableClDeposits />
          </div>
        </template>
      </BcTabList>
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

:global(.modal-controller) {
  max-width: 100%;
  width: 460px;
}

.overview {
  margin-bottom: var(--padding-large);
}

.dashboard-tab-view {
  margin-top: var(--padding-large);

  :deep(.dashboard-tab-panels) {
    min-height: 699px;
  }
}

.down_icon {
  width: 100%;
  height: 28px;
}
</style>
