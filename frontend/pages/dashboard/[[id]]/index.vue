<script setup lang="ts">
import {
  DashboardCreationController, DashboardTableBlocks, DashboardTableEmpty, DashboardTableRewards, DashboardTableSummary,
  DashboardTableWithdrawals,
} from '#components'
import type { GuestDashboard } from '~/types/dashboard'
import {
  isGuestDashboardKey, isSharedDashboardKey,
} from '~/utils/dashboard/key'
import type { HashTabs } from '~/types/hashTabs'

const {
  isLoggedIn,
} = useUserStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { t: $t } = useTranslation()

const tabs: HashTabs = [
  {
    component: DashboardTableSummary,
    icon: 'chart-line-up',
    key: 'summary',
    title: $t('dashboard.validator.tabs.summary'),
  },
  {
    component: DashboardTableRewards,
    icon: 'cubes',
    key: 'rewards',
    title: $t('dashboard.validator.tabs.rewards'),
  },
  {
    component: DashboardTableBlocks,
    icon: 'cube',
    key: 'blocks',
    title: $t('dashboard.validator.tabs.blocks'),

  },
  {
    component: DashboardTableEmpty,
    disabled: !showInDevelopment,
    icon: 'fire',
    key: 'heatmap',
    title: $t('dashboard.validator.tabs.heatmap'),
  },
  {
    icon: 'wallet',
    key: 'deposits',
    title: $t('dashboard.validator.tabs.deposits'),
  },
  {
    component: DashboardTableWithdrawals,
    icon: 'money-bill',
    key: 'withdrawals',
    title: $t('dashboard.validator.tabs.withdrawals'),
  },
]

const {
  dashboardKey,
  setDashboardKey,
} = useDashboardKeyProvider('validator')

const userDashboardStore = useUserDashboardStore()
const {
  getDashboardLabel,
  refreshDashboards,
  updateGuestDashboardKey,
} = userDashboardStore

const {
  cookieDashboards,
  dashboards,
} = storeToRefs(userDashboardStore)

const seoTitle = computed(() => {
  return getDashboardLabel(dashboardKey.value, 'validator')
})

useBcSeo(seoTitle, true)

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  overview,
} = storeToRefs(validatorDashboardOverviewStore)
const {
  refreshOverview,
} = validatorDashboardOverviewStore

const {
  getProducts,
} = useProductsStore()

await useAsyncData('get_products', () => getProducts())

await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

const { error: validatorOverviewError } = await useAsyncData(
  'validator_overview',
  () => refreshOverview(dashboardKey.value),
  { watch: [ dashboardKey ] },
)
// when we run into an error loading a dashboard keep it here to prevent an infinity loop
const errorDashboardKeys: string[] = []
const setDashboardKeyIfNoError = (key: string) => {
  if (!errorDashboardKeys.includes(key)) {
    setDashboardKey(key)
  }
}
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
      // Some checks if we need to update the dashboard key or the guest dashboard
      let gd = dashboards.value?.validator_dashboards?.[0] as GuestDashboard
      const isGuest = isGuestDashboardKey(newKey)
      const isShared = isSharedDashboardKey(newKey)
      if (isShared) {
        return
      }
      if (newLoggedIn) {
        // if we are logged in and have no dashboard key we only want to switch
        //  to the first dashboard if it is a private one
        if (gd && gd.key === undefined) {
          setDashboardKeyIfNoError(gd.id.toString())
        }
      }
      else if (
        !newLoggedIn
        && gd
        && isGuest
        && (!gd.key || (gd.key ?? '') === (oldKey ?? ''))
      ) {
        // we got a new guest dashboard key but the old key matches the
        // stored dashboard - so we update the stored dashboard
        if (!errorDashboardKeys.includes(newKey)) {
          updateGuestDashboardKey('validator', newKey)
        }
        setDashboardKeyIfNoError(newKey ?? '')
      }
      else if (!newKey || !isGuest) {
        // trying to view a private dashboad but not logged in
        gd = cookieDashboards.value
          ?.validator_dashboards?.[0] as GuestDashboard
        setDashboardKeyIfNoError(gd?.key ?? '')
      }
    }
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <BcPageWrapper>
      <DashboardCreationController
        class="panel-controller"
        :display-mode="'panel'"
        :initially-visible="true"
      />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-mode="'modal'"
    />
    <BcPageWrapper>
      <template #banner>
        <BcNotificationBanner
          v-if=" overview?.is_above_effective_balance_limit"
          :title="$t('dashboard.subsciprion_limit_reached_title')"
        >
          <BcTranslation
            keypath="dashboard.subsciprion_limit_reached.template"
            linkpath="dashboard.subsciprion_limit_reached._link"
            to="/pricing"
          />
        </BcNotificationBanner>
      </template>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardControls :dashboard-title="overview?.name" />
        <DashboardValidatorOverview class="overview" />
      </template>
      <DashboardSharedDashboardModal />
      <div>
        <DashboardSlotViz />
      </div>
      <BcTabList
        :tabs
        default-tab="summary"
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
      >
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits />
            <BcIcon
              name="arrow-down"
              class="down_icon"
            />
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
