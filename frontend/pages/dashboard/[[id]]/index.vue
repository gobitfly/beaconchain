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

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const {
  dashboardKey, setDashboardKey,
} = useDashboardKeyProvider('validator')

const { t: $t } = useTranslation()
const validatorDashboard = useValidatorDashboard()
const { isLoggedIn } = useUserStore()

// tick
const { networkInfo } = useNetworkStore()
const { secondsPerSlot = 12 } = networkInfo.value
const {
  resetTick, tick,
} = useInterval(secondsPerSlot)

// user dashboards
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

// dashboard page
const validatorDashboardStore = useValidatorDashboardStore()
const {
  initializeOverviewData, resetOverviewData,
} = validatorDashboardStore
const {
  populatedGroups,
} = storeToRefs(validatorDashboardStore)

// dashboard page title
const seoTitle = computed(() => getDashboardLabel(dashboardKey.value, 'validator'),
)
useBcSeo(seoTitle, true)

// initialize page with fresh data
await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

const { handleLoginOrKeyChange } = useLoginAndKeyChange()

const {
  isLoadingOverview, overview, refreshOverview,
} = useOverview()
const {
  isLoadingSlotViz,
  refetchingSlotVizData,
  refreshSlotViz,
  slotVizData,
  slotVizSelectedGroups,
} = useSlotViz()
const {
  activeTab,
  refreshActiveTab,
  tabs,
} = useDashboardTabs()
const {
  isLoadingSummary,
  resetSummaryQueryParams,
  resetSummaryTimeframe,
  summary,
  summaryQueryParams,
  summaryTimeframe,
} = useTableDataSummary()
const {
  isLoadingRewards, resetRewardsQueryParams, rewards, rewardsQueryParams,
} = useTableDataRewards()
const {
  blocks, blocksQueryParams, isLoadingBlocks, resetBlocksQueryParams,
} = useTableDataBlocks()
const {
  clDeposits,
  clDepositsQueryParams,
  isLoadingClDeposits,
  isLoadingTotalClDeposits,
  refreshTotalClDeposits,
  resetClDepositsQueryParams,
  totalClDeposits,
} = useTableDataClDeposits()
const {
  elDeposits,
  elDepositsQueryParams,
  isLoadingElDeposits,
  isLoadingTotalElDeposits,
  refreshTotalElDeposits,
  resetElDepositsQueryParams,
  totalElDeposits,
} = useTableDataElDeposits()
const {
  isLoadingTotalWithdrawals,
  isLoadingWithdrawals,
  refreshTotalWithdrawals,
  resetWidthdrawalsQueryParams,
  totalWithdrawals,
  withdrawals,
  withdrawalsQueryParams,
} = useTableDataWithdrawals()
const { showDashboardCreationModal } = useDashboardCreationModal()

// initial run
handleLoginOrKeyChange(dashboardKey.value, dashboardKey.value, isLoggedIn.value)

// logic specific to this component
function useDashboardCreationModal() {
  const dashboardCreationControllerModal
    = ref<typeof DashboardCreationController>()
  const showDashboardCreationModal = () => {
    dashboardCreationControllerModal.value?.show()
  }

  return { showDashboardCreationModal }
}
function useDashboardTabs() {
  const activeTab = ref()

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
    switch (activeTab.value) {
      case 'blocks':
        resetBlocksQueryParams()
        return
      case 'deposits':
        resetClDepositsQueryParams()
        resetElDepositsQueryParams()
        refreshTotalClDeposits()
        refreshTotalElDeposits()
        return
      case 'rewards':
        resetRewardsQueryParams()
        return
      case 'summary':
        resetSummaryQueryParams()
        resetSummaryTimeframe()
        return
      case 'withdrawals':
        resetWidthdrawalsQueryParams()
        refreshTotalWithdrawals()
        return
    }
  }

  watch(
    activeTab,
    () => {
      refreshActiveTab()
    },
  )

  return {
    activeTab,
    refreshActiveTab,
    tabs,
  }
}
function useLoginAndKeyChange() {
  const errorDashboardKeys: string[] = []

  const setDashboardKeyIfNoError = (key: string) => {
    if (!errorDashboardKeys.includes(key)) {
      setDashboardKey(key)
    }
  }
  const handleLoginOrKeyChange = (
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
    // to the first dashboard if it is a private one
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

  watch(
    [
      dashboardKey,
      isLoggedIn,
    ],
    ([
      newKey,
      newLoggedIn,
    ], [ oldKey ]) => {
      handleLoginOrKeyChange(oldKey, newKey, newLoggedIn)
      if (newKey) {
        refreshAll()
      }
    },
  )

  return { handleLoginOrKeyChange }
}
function useOverview() {
  const {
    data: overview,
    refresh: refreshOverview,
    status: overviewDataStatus,
  } = useAsyncData('validator_dashboard_overview', () => validatorDashboard.fetchOverview(dashboardKey.value))
  const isLoadingOverview = computed(() => isLoading(overviewDataStatus))

  watch(
    overview,
    (newOverview) => {
      if (newOverview?.data) {
        initializeOverviewData(newOverview.data)
      }
    },
    { immediate: true },
  )

  return {
    isLoadingOverview, overview, refreshOverview,
  }
}
function useSlotViz() {
  const slotVizSelectedGroups = ref<number[]>([])
  const refetchingSlotVizData = ref(false)

  const {
    data: slotVizData,
    refresh: refreshSlotViz,
    status: slotVizDataStatus,
  } = useAsyncData('validator_dashboard_slot_viz',
    () => validatorDashboard.fetchSlotViz(dashboardKey.value, slotVizSelectedGroups.value))
  const isLoadingSlotViz = computed(() => isLoading(slotVizDataStatus))

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

  return {
    isLoadingSlotViz,
    refetchingSlotVizData,
    refreshSlotViz,
    slotVizData,
    slotVizSelectedGroups,
  }
}
const defaultPageSize = 10

function isLoading(status: Ref<AsyncDataRequestStatus>) {
  return status.value === 'pending'
}
function useTableDataBlocks() {
  const defaultBlocksQueryParams: TableQueryParams = {
    limit: defaultPageSize,
    sort: 'slot:desc',
  }
  const blocksQueryParams = ref<TableQueryParams>(defaultBlocksQueryParams)

  const {
    data: blocks,
    status: statusBlocks,
  } = useAsyncData('validator_dashboard_blocks', () => validatorDashboard.fetchBlocks(dashboardKey.value, blocksQueryParams.value), {
    immediate: false,
    watch: [ blocksQueryParams ],
  })
  const isLoadingBlocks = isLoading(statusBlocks)

  const resetBlocksQueryParams = () => blocksQueryParams.value = { ...defaultBlocksQueryParams }

  return {
    blocks, blocksQueryParams, isLoadingBlocks, resetBlocksQueryParams,
  }
}

function useTableDataClDeposits() {
  const defaultClDepositsQueryParams: TableQueryParams = {
    limit: 5,
  }
  const clDepositsQueryParams = ref<TableQueryParams>(defaultClDepositsQueryParams)

  const {
    data: clDeposits,
    status: statusClDeposits,
  } = useAsyncData('validator_dashboard_cl_deposits', () => validatorDashboard.fetchClDeposits(dashboardKey.value, clDepositsQueryParams.value), {
    immediate: false,
    watch: [ clDepositsQueryParams ],
  })
  const {
    data: totalClDeposits,
    refresh: refreshTotalClDeposits,
    status: statusTotalClDeposits,
  } = useAsyncData('validator_dashboard_total_cl_deposits', () => validatorDashboard.fetchTotalClDeposits(dashboardKey.value), {
    immediate: false,
  })

  const isLoadingClDeposits = isLoading(statusClDeposits)
  const isLoadingTotalClDeposits = isLoading(statusTotalClDeposits)

  const resetClDepositsQueryParams = () => clDepositsQueryParams.value = { ...defaultClDepositsQueryParams }

  return {
    clDeposits,
    clDepositsQueryParams,
    isLoadingClDeposits,
    isLoadingTotalClDeposits,
    refreshTotalClDeposits,
    resetClDepositsQueryParams,
    totalClDeposits,
  }
}
function useTableDataElDeposits() {
  const defaultElDepositsQueryParams: TableQueryParams = {
    limit: 5,
  }
  const elDepositsQueryParams = ref<TableQueryParams>(defaultElDepositsQueryParams)

  const {
    data: elDeposits,
    status: statusElDeposits,
  } = useAsyncData('validator_dashboard_el_deposits', () => validatorDashboard.fetchElDeposits(dashboardKey.value, elDepositsQueryParams.value), {
    immediate: false,
    watch: [ elDepositsQueryParams ],
  })
  const {
    data: totalElDeposits,
    refresh: refreshTotalElDeposits,
    status: statusTotalElDeposits,
  } = useAsyncData('validator_dashboard_total_el_deposits', () => validatorDashboard.fetchTotalElDeposits(dashboardKey.value), {
    immediate: false,
  })

  const isLoadingTotalElDeposits = isLoading(statusTotalElDeposits)
  const isLoadingElDeposits = isLoading(statusElDeposits)

  const resetElDepositsQueryParams = () => elDepositsQueryParams.value = { ...defaultElDepositsQueryParams }

  return {
    elDeposits,
    elDepositsQueryParams,
    isLoadingElDeposits,
    isLoadingTotalElDeposits,
    refreshTotalElDeposits,
    resetElDepositsQueryParams,
    totalElDeposits,
  }
}
function useTableDataRewards() {
  const defaultRewardsQueryParams: TableQueryParams = {
    limit: defaultPageSize,
    sort: 'epoch:desc',
  }
  const rewardsQueryParams = ref<TableQueryParams>(defaultRewardsQueryParams)

  const {
    data: rewards,
    status: statusRewards,
  } = useAsyncData('validator_dashboard_rewards', () => validatorDashboard.fetchRewards(dashboardKey.value, rewardsQueryParams.value), {
    immediate: false,
    watch: [ rewardsQueryParams ],
  })

  const isLoadingRewards = isLoading(statusRewards)

  const resetRewardsQueryParams = () => rewardsQueryParams.value = { ...defaultRewardsQueryParams }

  const createFutureRewardRow = (latestEpoch?: SlotVizEpoch): VDBRewardsTableRow => {
    return {
      duty: {
        attestation: latestEpoch?.slots?.find(s => s.attestations) ? 0 : undefined,
        proposal: latestEpoch?.slots?.find(s => s.proposal) ? 0 : undefined,
        slashing: latestEpoch?.slots?.find(s => s.slashing) ? 0 : undefined,
        sync: latestEpoch?.slots?.find(s => s.sync) ? 0 : undefined,
      },
      epoch: latestEpoch?.epoch || 0,
      group_id: DAHSHBOARDS_NEXT_EPOCH_ID,
      reward: {
        cl: '0', el: '0',
      },
    }
  }

  const rewardsWithFutureRow = computed(() => {
    if (!rewards.value?.data) return null

    const isFirstPage = !rewards.value?.paging?.prev_cursor
    const dataEpoch = rewards.value?.data[0].epoch
    const latestEpoch = slotVizData.value?.[0].epoch ?? 0

    if (!isFirstPage || !slotVizData || slotVizData.value?.length === 0 || latestEpoch <= dataEpoch) {
      // Already up to date or not on the first page
      return rewards.value
    }

    // Otherwise, create and add future row from slot visualization data
    const futureRewardRow = createFutureRewardRow(slotVizData.value?.[0])

    return {
      data: [
        futureRewardRow,
        ...rewards.value.data,
      ],
      paging: rewards.value.paging,
    }
  })

  return {
    isLoadingRewards,
    resetRewardsQueryParams,
    rewards: rewardsWithFutureRow,
    rewardsQueryParams,
  }
}
function useTableDataSummary() {
  const defaultSummaryQueryParams: TableQueryParams = {
    limit: defaultPageSize,
    sort: 'efficiency:desc',
  }
  const defaultSummaryTimeframe: SummaryTimeFrame = 'last_24h'
  const summaryTimeframe = ref<SummaryTimeFrame>(defaultSummaryTimeframe)
  const summaryQueryParams = ref<TableQueryParams>(defaultSummaryQueryParams)
  const {
    data: summary,
    status: statusSummary,
  } = useAsyncData('validator_dashboard_summary', () => validatorDashboard.fetchSummary(dashboardKey.value, summaryTimeframe.value, summaryQueryParams.value), {
    immediate: false,
    watch: [
      summaryQueryParams,
      summaryTimeframe,
    ],
  })
  const isLoadingSummary = isLoading(statusSummary)

  const resetSummaryQueryParams = () => summaryQueryParams.value = { ...defaultSummaryQueryParams }
  const resetSummaryTimeframe = () => summaryQueryParams.value = { ...defaultSummaryQueryParams }

  return {
    isLoadingSummary,
    resetSummaryQueryParams,
    resetSummaryTimeframe,
    summary,
    summaryQueryParams,
    summaryTimeframe,
  }
}

function useTableDataWithdrawals() {
  const defaultwithdrawalsQueryParams: TableQueryParams = {
    limit: defaultPageSize,
    sort: 'slot:desc',
  }
  const withdrawalsQueryParams = ref<TableQueryParams>(defaultwithdrawalsQueryParams)

  const {
    data: withdrawals,
    status: statusWithdrawals,
  } = useAsyncData('validator_dashboard_withdrawals', () => validatorDashboard.fetchWithdrawals(dashboardKey.value, withdrawalsQueryParams.value), {
    immediate: false,
    watch: [ withdrawalsQueryParams ],
  })
  const {
    data: totalWithdrawals,
    refresh: refreshTotalWithdrawals,
    status: statusTotalWithdrawals,
  } = useAsyncData('validator_dashboard_total_withdrawals', () => validatorDashboard.fetchTotalWithdrawals(dashboardKey.value), {
    immediate: false,
  })

  const isLoadingWithdrawals = isLoading(statusWithdrawals)
  const isLoadingTotalWithdrawals = isLoading(statusTotalWithdrawals)

  const resetWidthdrawalsQueryParams = () => withdrawalsQueryParams.value = { ...defaultwithdrawalsQueryParams }

  return {
    isLoadingTotalWithdrawals,
    isLoadingWithdrawals,
    refreshTotalWithdrawals,
    resetWidthdrawalsQueryParams,
    totalWithdrawals,
    withdrawals,
    withdrawalsQueryParams,
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
        <DashboardHeader @show-creation="showDashboardCreationModal" />
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
        v-if="overview && slotVizData"
        v-model:selected-groups="slotVizSelectedGroups"
        :epochs-data="slotVizData"
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
            v-model:query="summaryQueryParams"
            v-model:time-frame="summaryTimeframe"
            :data="summary?.data"
            :paging="summary?.paging"
            :is-loading="isLoadingSummary"
          />
        </template>
        <template #tab-panel-rewards>
          <DashboardTableRewards
            v-model:query="rewardsQueryParams"
            :data="rewards?.data"
            :paging="rewards?.paging"
            :is-loading="isLoadingRewards"
          />
        </template>
        <template #tab-panel-blocks>
          <DashboardTableBlocks
            v-model:query="blocksQueryParams"
            :data="blocks?.data"
            :paging="blocks?.paging"
            :is-loading="isLoadingBlocks"
          />
        </template>
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits
              v-model:query="elDepositsQueryParams"
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
              v-model:query="clDepositsQueryParams"
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
            v-model:query="withdrawalsQueryParams"
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
