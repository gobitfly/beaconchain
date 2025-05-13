<script setup lang="ts">
import {
  DashboardCreationController, DashboardTableBlocks, DashboardTableEmpty, DashboardTableRewards, DashboardTableSummary,
} from '#components'
import type { GuestDashboard } from '~/types/dashboard'
import {
  isGuestDashboardKey, isSharedDashboardKey,
} from '~/utils/dashboard/key'
import type { HashTabs } from '~/types/hashTabs'
import type { TableQueryParams } from '~/types/datatable'

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
    icon: 'money-bill',
    key: 'withdrawals',
    title: $t('dashboard.validator.tabs.withdrawals'),
  },
  {
    icon: 'hand-holding-hand',
    key: 'consolidations',
    title: $t('dashboard.validator.tabs.consolidations'),
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
  hasValidators,
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

const dashboardData = useDashboardData()

// Execution Layer Deposits data
const elDepositsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: elDepositsData,
  refresh: refreshElDepositsData,
  status: elDepositsDataStatus,
} = useAsyncData('el_deposits', () => {
  return Promise.all([
    dashboardData.fetchELDeposits(
      dashboardKey.value,
      elDepositsQueryParams.value,
    ),
    dashboardData.fetchELDpositsTotalAmount(
      dashboardKey.value,
      { search: elDepositsQueryParams.value.search },
    ),
  ])
},
{
  immediate: false,
  watch: [ elDepositsQueryParams ],
})
const elDeposits = computed(() => {
  return elDepositsData.value?.[0]
})
const elDepositsTotalAmount = computed(() => elDepositsData.value?.[1])

// Consensus Layer Deposits data
const clDepositsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: clDepositsData,
  refresh: refreshClDepositsData,
  status: clDepositsDataStatus,
} = useAsyncData('cl_deposits', () => {
  return Promise.all([
    dashboardData.fetchClDeposits(
      dashboardKey.value,
      clDepositsQueryParams.value,
    ),
    dashboardData.fetchClDpositsTotalAmount(
      dashboardKey.value,
      { search: clDepositsQueryParams.value.search },
    ),
  ])
},
{
  immediate: false,
  watch: [ clDepositsQueryParams ],
})
const clDeposits = computed(() => {
  return clDepositsData.value?.[0]
})
const clDepositsTotalAmount = computed(() => clDepositsData.value?.[1])

// Execution Layer Withdrawals data
const elWithdrawalsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: elWithdrawalsData,
  refresh: refreshElWithdrawalsData,
  status: elWithdrawalsDataStatus,
} = useAsyncData('el_withdrawals', () => {
  return Promise.all([
    dashboardData.fetchElWithdrawals(
      dashboardKey.value,
      elWithdrawalsQueryParams.value,
    ),
    dashboardData.fetchElWithdrawalsTotalAmount(
      dashboardKey.value,
      { search: elWithdrawalsQueryParams.value.search },
    ),
  ])
},
{
  immediate: false,
  watch: [ elWithdrawalsQueryParams ],
})
const elWithdrawals = computed(() => {
  return elWithdrawalsData.value?.[0]
})
const elWithdrawalsTotalAmount = computed(() => elWithdrawalsData.value?.[1])

// Consensus Layer Withdrawals data
const clWithdrawalsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: clWithdrawalsData,
  refresh: refreshClWithdrawalsData,
  status: clWithdrawalsDataStatus,
} = useAsyncData('cl_withdrawals', () => {
  return Promise.all([
    dashboardData.fetchClWithdrawals(
      dashboardKey.value,
      clWithdrawalsQueryParams.value,
    ),
    dashboardData.fetchClWithdrawalsTotalAmount(
      dashboardKey.value,
      { search: clWithdrawalsQueryParams.value.search },
    ),
  ])
},
{
  immediate: false,
  watch: [ clWithdrawalsQueryParams ],
})
const clWithdrawals = computed(() => {
  return clWithdrawalsData.value?.[0]
})
const clWithdrawalsTotalAmount = computed(() => clWithdrawalsData.value?.[1])

// Execution Layer Consolidations data
const elConsolidationsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: elConsolidationsData,
  refresh: refreshElConsolidationsData,
  status: elConsolidationsDataStatus,
} = useAsyncData('el_consolidations', () => {
  return dashboardData.fetchElConsolidations(
    dashboardKey.value,
    elConsolidationsQueryParams.value,
  )
},
{
  immediate: false,
  watch: [ elConsolidationsQueryParams ],
})
const elConsolidations = computed(() => {
  return elConsolidationsData.value || undefined
})

// Consensus Layer Consolidations data
const clConsolidationsQueryParams = ref<TableQueryParams>({
  limit: 5,
  sort: 'timestamp:desc',
})
const {
  data: clConsolidationsData,
  refresh: refreshClConsolidationsData,
  status: clConsolidationsDataStatus,
} = useAsyncData('cl_consolidations', () => {
  return dashboardData.fetchClConsolidations(
    dashboardKey.value,
    clConsolidationsQueryParams.value,
  )
},
{
  immediate: false,
  watch: [ clConsolidationsQueryParams ],
})
const clConsolidations = computed(() => {
  return clConsolidationsData.value || undefined
})

// tabs
const route = useRoute()

const activeTab = computed(() => route.hash)

const refreshActiveTab = () => {
  if (!hasValidators.value) return

  switch (activeTab.value) {
    case '#consolidations':
      refreshElConsolidationsData()
      refreshClConsolidationsData()
      break
    case '#deposits':
      refreshElDepositsData()
      refreshClDepositsData()
      break
    case '#withdrawals':
      refreshElWithdrawalsData()
      refreshClWithdrawalsData()
      break
  }
}

watch([
  activeTab,
  overview,
],
() => {
  refreshActiveTab()
},
{
  immediate: true,
},
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

      <DashboardSlotViz />

      <BcTabList
        :tabs
        default-tab="summary"
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
      >
        <template #tab-panel-deposits>
          <DashboardTableElDeposits
            v-model:query="elDepositsQueryParams"
            :el-deposits
            :el-deposits-total-amount
            :is-loading="elDepositsDataStatus === 'pending'"
          />
          <BcIcon
            name="arrow-down"
            class="down_icon"
          />
          <DashboardTableClDeposits
            v-model:query="clDepositsQueryParams"
            :cl-deposits
            :cl-deposits-total-amount
            :is-loading="clDepositsDataStatus === 'pending'"
          />
        </template>
        <template #tab-panel-withdrawals>
          <DashboardTableElWithdrawals
            v-model:query="elWithdrawalsQueryParams"
            :el-withdrawals
            :el-withdrawals-total-amount
            :is-loading="elWithdrawalsDataStatus === 'pending'"
          />
          <BcIcon
            name="arrow-down"
            class="down_icon"
          />
          <DashboardTableClWithdrawals
            v-model:query="clWithdrawalsQueryParams"
            :cl-withdrawals
            :cl-withdrawals-total-amount
            :is-loading="clWithdrawalsDataStatus === 'pending'"
          />
        </template>
        <template #tab-panel-consolidations>
          <DashboardTableElConsolidations
            v-model:query="elConsolidationsQueryParams"
            :el-consolidations
            :is-loading="elConsolidationsDataStatus === 'pending'"
          />
          <BcIcon
            name="arrow-down"
            class="down_icon"
          />
          <DashboardTableClConsolidations
            v-model:query="clConsolidationsQueryParams"
            :cl-consolidations
            :is-loading="clConsolidationsDataStatus === 'pending'"
          />
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
