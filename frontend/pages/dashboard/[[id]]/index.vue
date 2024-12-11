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
  DashboardCreationController,
  DashboardTableBlocks,
  DashboardTableEmpty,
  DashboardTableRewards,
  DashboardTableSummary,
  DashboardTableWithdrawals,
} from '#components'
import {
  DAHSHBOARDS_NEXT_EPOCH_ID,
  type DashboardKey,
  type GuestDashboard,
} from '~/types/dashboard'
import {
  isGuestDashboardKey, isSharedDashboardKey,
} from '~/utils/dashboard/key'
import type { HashTabs } from '~/types/hashTabs'
import type { VDBRewardsTableRow } from '~/types/api/validator_dashboard'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import type { TableQueryParams } from '~/types/datatable'
import type { SummaryTimeFrame } from '~/types/dashboard/summary'
import type { AsyncDataRequestStatus } from '#app'

const { isLoggedIn } = useUserStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { t: $t } = useTranslation()
const { networkInfo } = useNetworkStore()
const validatorDashboard = useValidatorDashboard()

const {
  dashboardKey, setDashboardKey,
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
const validatorDashboardStore = useValidatorDashboardStore()
const {
  initializeOverviewData, resetOverviewData,
} = validatorDashboardStore
const {
  populatedGroups,
} = storeToRefs(validatorDashboardStore)

const seoTitle = computed(() => getDashboardLabel(dashboardKey.value, 'validator'),
)
useBcSeo(seoTitle, true)

const dashboardCreationControllerModal
  = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog() {
  dashboardCreationControllerModal.value?.show()
}

const useIsLoading = (status: Ref<AsyncDataRequestStatus>) => {
  return computed(() => status.value === 'pending')
}

await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

// LOGIN AND DASHBOARD KEY CHANGE
const errorDashboardKeys: string[] = []

const setDashboardKeyIfNoError = (key: string) => {
  if (!errorDashboardKeys.includes(key)) {
    setDashboardKey(key)
  }
}
const handleKeyOrLoginChange = (
  oldKey: DashboardKey,
  newKey: DashboardKey,
  newLoggedIn: boolean,
) => {
  if (newLoggedIn && newKey) {
    return
  }
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
    gd
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

handleKeyOrLoginChange(dashboardKey.value, dashboardKey.value, isLoggedIn.value) // initial run

watch(
  [
    dashboardKey,
    isLoggedIn,
  ],
  ([
    newKey,
    newLoggedIn,
  ], [ oldKey ]) => {
    handleKeyOrLoginChange(oldKey, newKey, newLoggedIn)
    if (newKey) {
      refreshAll()
    }
  },
)

// OVERVIEW
const {
  data: overview,
  refresh: refreshOverview,
  status: overviewDataStatus,
} = useAsyncData('validator_dashboard_overview', () => validatorDashboard.fetchOverview(dashboardKey.value))
const isLoadingOverview = computed(() => useIsLoading(overviewDataStatus).value)
watch(
  overview,
  (newOverview) => {
    if (newOverview?.data) {
      initializeOverviewData(newOverview.data)
    }
  },
  { immediate: true },
)

// SLOTVIZ
const { secondsPerSlot = 12 } = networkInfo.value
const {
  resetTick, tick,
} = useInterval(secondsPerSlot)

const slotVizSelectedGroups = ref<number[]>([])
const refetchingSlotVizData = ref(false)

const {
  data: dataSlotViz,
  refresh: refreshSlotViz,
  status: slotVizDataStatus,
} = useAsyncData('validator_dashboard_slot_viz', () => validatorDashboard.fetchSlotViz(dashboardKey.value, slotVizSelectedGroups.value))
const isLoadingSlotViz = computed(() => useIsLoading(slotVizDataStatus).value)

const hasNoGroupsOrAllSelected = (groups: number[]) => {
  return groups.length === 0 || groups.length === populatedGroups.value?.length
}

watch(
  tick,
  async () => {
    refetchingSlotVizData.value = true
    await refreshSlotViz()
    refetchingSlotVizData.value = false
  },
)
watch(
  slotVizSelectedGroups,
  (newSlotVizGroups, oldSlotVizGroups) => {
    if (hasNoGroupsOrAllSelected(newSlotVizGroups) && hasNoGroupsOrAllSelected(oldSlotVizGroups)) {
      // don't refresh redundantly if all groups are selected and were before
      return
    }
    resetTick()
    refreshSlotViz()
  },
)

// TABS SECTION
const defaultPageSize = 10
const defaultQueryRewards: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'epoch:desc',
}
const defaultTimeframeSummary: SummaryTimeFrame = 'last_24h'
const defaultQuerySummary: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'efficiency:desc',
}
const defaultQueryBlocks: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}
const defaultQueryClDeposits: TableQueryParams = {
  limit: 5,
}
const defaultQueryElDeposits: TableQueryParams = {
  limit: 5,
}
const defaultQueryWithdrawals: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}

const activeTab = ref<string>()
const querySummary = ref<TableQueryParams>(defaultQuerySummary)
const timeframeSummary = ref<SummaryTimeFrame>(defaultTimeframeSummary)
const queryRewards = ref<TableQueryParams>(defaultQueryRewards)
const queryBlocks = ref<TableQueryParams>(defaultQueryBlocks)
const queryClDeposits = ref<TableQueryParams>(defaultQueryClDeposits)
const queryElDeposits = ref<TableQueryParams>(defaultQueryElDeposits)
const queryWithdrawals = ref<TableQueryParams>(defaultQueryWithdrawals)

const {
  data: summary,
  status: statusSummary,
} = useAsyncData('validator_dashboard_summary', () => validatorDashboard.fetchSummary(dashboardKey.value, timeframeSummary.value, querySummary.value), {
  immediate: false,
  watch: [
    querySummary,
    timeframeSummary,
  ],
})
const {
  data: rewards,
  status: statusRewards,
} = useAsyncData('validator_dashboard_rewards', () => validatorDashboard.fetchRewards(dashboardKey.value, queryRewards.value), {
  immediate: false,
  watch: [ queryRewards ],
})
const {
  data: blocks,
  status: statusBlocks,
} = useAsyncData('validator_dashboard_blocks', () => validatorDashboard.fetchBlocks(dashboardKey.value, queryBlocks.value), {
  immediate: false,
  watch: [ queryBlocks ],
})
const {
  data: clDeposits,
  status: statusClDeposits,
} = useAsyncData('validator_dashboard_cl_deposits', () => validatorDashboard.fetchClDeposits(dashboardKey.value, queryClDeposits.value), {
  immediate: false,
  watch: [ queryClDeposits ],
})
const {
  data: totalClDeposits,
  refresh: refreshTotalClDeposits,
  status: statusTotalClDeposits,
} = useAsyncData('validator_dashboard_total_cl_deposits', () => validatorDashboard.fetchTotalClDeposits(dashboardKey.value), {
  immediate: false,
})
const {
  data: elDeposits,
  status: statusElDeposits,
} = useAsyncData('validator_dashboard_el_deposits', () => validatorDashboard.fetchElDeposits(dashboardKey.value, queryElDeposits.value), {
  immediate: false,
  watch: [ queryElDeposits ],
})
const {
  data: totalElDeposits,
  refresh: refreshTotalElDeposits,
  status: statusTotalElDeposits,
} = useAsyncData('validator_dashboard_total_el_deposits', () => validatorDashboard.fetchTotalElDeposits(dashboardKey.value), {
  immediate: false,
})
const {
  data: withdrawals,
  status: statusWithdrawals,
} = useAsyncData('validator_dashboard_withdrawals', () => validatorDashboard.fetchWithdrawals(dashboardKey.value, queryWithdrawals.value), {
  immediate: false,
  watch: [ queryWithdrawals ],
})
const {
  data: totalWithdrawals,
  refresh: refreshTotalWithdrawals,
  status: statusTotalWithdrawals,
} = useAsyncData('validator_dashboard_total_withdrawals', () => validatorDashboard.fetchTotalWithdrawals(dashboardKey.value), {
  immediate: false,
})

const isLoadingClDeposits = useIsLoading(statusClDeposits)
const isLoadingSummary = useIsLoading(statusSummary)
const isLoadingTotalClDeposits = useIsLoading(statusTotalClDeposits)
const isLoadingBlocks = useIsLoading(statusBlocks)
const isLoadingTotalElDeposits = useIsLoading(statusTotalElDeposits)
const isLoadingElDeposits = useIsLoading(statusElDeposits)
const isLoadingWithdrawals = useIsLoading(statusWithdrawals)
const isLoadingTotalWithdrawals = useIsLoading(statusTotalWithdrawals)
const isLoadingRewards = useIsLoading(statusRewards)

const tabs: HashTabs = [
  {
    icon: faChartLineUp,
    key: 'summary',
    title: $t('dashboard.validator.tabs.summary'),
  },
  {
    icon: faCubes,
    key: 'rewards',
    title: $t('dashboard.validator.tabs.rewards'),
  },
  {
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
    icon: faMoneyBill,
    key: 'withdrawals',
    title: $t('dashboard.validator.tabs.withdrawals'),
  },
]
const refreshActiveTab = () => {
  // always assign query as shallow copy to trigger watch
  switch (activeTab.value) {
    case 'blocks':
      queryBlocks.value = { ...defaultQueryBlocks }
      return
    case 'deposits':
      refreshTotalClDeposits()
      refreshTotalElDeposits()
      queryClDeposits.value = { ...defaultQueryClDeposits }
      queryElDeposits.value = { ...defaultQueryElDeposits }
      return
    case 'rewards':
      queryRewards.value = { ...queryRewards.value }
      return
    case 'summary':
      querySummary.value = { ...defaultQuerySummary }
      timeframeSummary.value = defaultTimeframeSummary
      return
    case 'withdrawals':
      refreshTotalWithdrawals()
      queryWithdrawals.value = { ...defaultQueryWithdrawals }
      return
  }
}
const refreshAll = () => {
  if (!dashboardKey.value) {
    resetOverviewData()
  }

  resetTick()
  refreshSlotViz()
  refreshOverview()
  refreshActiveTab()
}
// Helper function to create a "future row" for rewards from slot viz data
function createNextRewardRow(latestEpoch: SlotVizEpoch): VDBRewardsTableRow {
  return {
    duty: {
      attestation: latestEpoch.slots?.find(s => s.attestations) ? 0 : undefined,
      proposal: latestEpoch.slots?.find(s => s.proposal) ? 0 : undefined,
      slashing: latestEpoch.slots?.find(s => s.slashing) ? 0 : undefined,
      sync: latestEpoch.slots?.find(s => s.sync) ? 0 : undefined,
    },
    epoch: latestEpoch.epoch,
    group_id: DAHSHBOARDS_NEXT_EPOCH_ID,
    reward: {
      cl: '0', el: '0',
    },
  }
}
// rewards data with added "future row"
const getRewardsData = () => {
  const data = rewards.value?.data
  if (!data || data.length === 0) {
    return undefined
  }

  const isFirstPage = !rewards.value?.paging?.prev_cursor
  const slotVizData = dataSlotViz.value
  const dataEpoch = data[0].epoch
  const latestEpoch = slotVizData?.[0].epoch ?? 0

  if (!isFirstPage || !slotVizData || slotVizData.length === 0 || latestEpoch <= dataEpoch) {
    // Already up to date or not on the first page
    return data
  }

  // Add future row from slot visualization data
  const nextRewardRow = createNextRewardRow(slotVizData[0])

  return [
    nextRewardRow,
    ...data,
  ]
}

watch(
  activeTab,
  () => {
    if (isServerSide) {
      return // url hash (where tab is stored) can't be read on server
    }
    refreshActiveTab()
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
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardControls
          :dashboard-title="overview?.data.name"
          @dashboard-modified="refreshAll()"
        />
        <DashboardValidatorOverview
          class="overview"
          v-bind="overview?.data"
        />
      </template>
      <DashboardSharedDashboardModal />

      <DashboardSlotViz
        v-if="overview && dataSlotViz"
        v-model:selected-groups="slotVizSelectedGroups"
        :epochs-data="dataSlotViz"
        :overview-data="overview?.data"
        :is-loading="isLoadingSlotViz || isLoadingOverview"
        :refetching-slot-viz-data
        :timestamp="tick"
      />

      <BcTabList
        :tabs
        :default-tab="'summary'"
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
        @changed-tab="(value) => activeTab = value"
      >
        <template #tab-panel-summary>
          <DashboardTableSummary
            v-model:query="querySummary"
            v-model:time-frame="timeframeSummary"
            :data="summary?.data"
            :paging="summary?.paging"
            :is-loading="isLoadingSummary"
          />
        </template>
        <template #tab-panel-rewards>
          <DashboardTableRewards
            v-model:query="queryRewards"
            :data="getRewardsData()"
            :paging="rewards?.paging"
            :is-loading="isLoadingRewards"
          />
        </template>
        <template #tab-panel-blocks>
          <DashboardTableBlocks
            v-model:query="queryBlocks"
            :data="blocks?.data"
            :paging="blocks?.paging"
            :is-loading="isLoadingBlocks"
          />
        </template>
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits
              v-model:query="queryElDeposits"
              :data="elDeposits?.data"
              :paging="elDeposits?.paging"
              :is-loading="isLoadingElDeposits"
              :is-loading-total="isLoadingTotalElDeposits"
              :total-amount="totalElDeposits?.data.total_amount"
            />
            <FontAwesomeIcon
              :icon="faArrowDown"
              class="down_icon"
            />
            <DashboardTableClDeposits
              v-model:query="queryClDeposits"
              :data="clDeposits?.data"
              :paging="clDeposits?.paging"
              :is-loading="isLoadingClDeposits"
              :is-loading-total="isLoadingTotalClDeposits"
              :total-amount="totalClDeposits?.data.total_amount"
            />
          </div>
        </template>
        <template #tab-panel-withdrawals>
          <DashboardTableWithdrawals
            v-model:query="queryWithdrawals"
            :data="withdrawals?.data"
            :paging="withdrawals?.paging"
            :is-loading="isLoadingWithdrawals"
            :is-loading-total="isLoadingTotalWithdrawals"
            :total-amount="totalWithdrawals?.data.total_amount"
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
