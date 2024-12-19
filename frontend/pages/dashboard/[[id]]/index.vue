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
import type {
  VDBOverviewData,
  VDBRewardsTableRow,
  VDBSummaryTableRow,
} from '~/types/api/validator_dashboard'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import type { TableQueryParams } from '~/types/datatable'
import type { Paging } from '~/types/api/common'
import type { SummaryTimeFrame } from '~/types/dashboard/summary'
import { useTableFetcher } from '~/composables/useTableFetcher'
import { useDataFetcher } from '~/composables/useDataFetcher'

const { isLoggedIn } = useUserStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { t: $t } = useTranslation()

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
await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

const seoTitle = computed(() => {
  return getDashboardLabel(dashboardKey.value, 'validator')
})
useBcSeo(seoTitle, true)

const dashboardCreationControllerModal
  = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog() {
  dashboardCreationControllerModal.value?.show()
}

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

const handleError = (e: any) => {
  if (!e) {
    return
  }
  if (e.statusCode === 404) {
    // TODO: show that the dashboard does not exist
    return
  }
  throw e
}

const validatorDashboardStore = useValidatorDashboardStore()
const service = useValidatorDashboard()

const dataOverview = ref<VDBOverviewData>()
const dataSlotViz = ref<SlotVizEpoch[]>()
const {
  fetchOverview,
  fetchSlotViz,
} = service

const refreshSlotViz = (groupIds: number[]) => {
  fetchSlotViz(dashboardKey.value, groupIds)
    .then((data) => {
      dataSlotViz.value = data
    })
    .catch((e) => {
      handleError(e)
    })
}

const defaultPageSize = 10

const defaultQuerySummary: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'efficiency:desc',
}
const dataSummary = ref<VDBSummaryTableRow[]>()
const pagingSummary = ref<Paging>()
const isLoadingSummary = ref(false)
const querySummary = ref<TableQueryParams>(defaultQuerySummary)
const timeframeSummary = ref<SummaryTimeFrame>('last_24h')
const refreshSummary = (timeframe: SummaryTimeFrame, query: TableQueryParams) => {
  isLoadingSummary.value = true
  querySummary.value = query
  timeframeSummary.value = timeframe
  service.fetchSummary(dashboardKey.value, timeframe, query)
    .then((fetchedResult) => {
      dataSummary.value = fetchedResult?.data
      pagingSummary.value = fetchedResult?.paging
    })
    .catch((e) => {
      handleError(e)
    })
    .finally(() => {
      isLoadingSummary.value = false
    })
}

const defaultQueryRewards: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'epoch:desc',
}
const rewards = useTableFetcher(service.fetchRewards, defaultQueryRewards, handleError)

// Helper function to create a "future row" for rewards data
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

// rewards data with a potential "future row"
const dataRewards = computed(() => {
  const data = rewards.data.value
  if (!data || data.length === 0) {
    return undefined
  }

  const isFirstPage = !rewards.paging.value?.prev_cursor
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
})

const defaultQueryBlocks: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}
const blocks = useTableFetcher(service.fetchBlocks, defaultQueryBlocks, handleError)

const defaultQueryClDeposits: TableQueryParams = {
  limit: 5,
}
const clDeposits = useTableFetcher(service.fetchClDeposits, defaultQueryClDeposits, handleError)
const totalClDeposits = useDataFetcher(service.fetchTotalClDeposits, handleError)
const defaultQueryElDeposits: TableQueryParams = {
  limit: 5,
}
const elDeposits = useTableFetcher(service.fetchElDeposits, defaultQueryElDeposits, handleError)
const totalElDeposits = useDataFetcher(service.fetchTotalElDeposits, handleError)
const defaultQueryWithdrawals: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}
const withdrawals = useTableFetcher(service.fetchWithdrawals, defaultQueryWithdrawals, handleError)
const totalWithdrawals = useDataFetcher(service.fetchTotalWithdrawals, handleError)

// fetches all data for the dashboard (overview, slot viz)
const fetchOverviewAndSlotViz = () => {
  if (dashboardKey.value) { // valid dashboard key -> fetch all data
    return Promise.allSettled([
      fetchOverview(dashboardKey.value),
      fetchSlotViz(dashboardKey.value),
    ])
  }
  if (!isLoggedIn.value) { // implies empty guest dashboard -> only fetch slot viz
    return Promise.allSettled([
      undefined,
      fetchSlotViz(dashboardKey.value),
    ])
  }
  return Promise.allSettled([ // implies logged-in user with no dashboards -> fetch nothing
    undefined,
    undefined,
  ])
}
const handleAllData = (
  fetchedOverviewData: PromiseFulfilledResult<undefined | VDBOverviewData>
    | PromiseRejectedResult,
  fetchedSlotVizData: PromiseFulfilledResult<SlotVizEpoch[] | undefined>
    | PromiseRejectedResult,
) => {
  if (isRejected(fetchedOverviewData)) {
    handleError(fetchedOverviewData.reason)
    return
  }
  if (isRejected(fetchedSlotVizData)) {
    handleError(fetchedSlotVizData.reason)
    return
  }
  dataOverview.value = fetchedOverviewData.value
  dataSlotViz.value = fetchedSlotVizData.value
  if (fetchedOverviewData.value) {
    validatorDashboardStore.setByOverviewData(fetchedOverviewData.value)
  }
}
function isRejected<T>(p: PromiseSettledResult<T>): p is PromiseRejectedResult {
  return p.status === 'rejected'
}
// init SSR data
const {
  data,
} = await useAsyncData('complete_validator_dashboard_fetch', async () => {
  return fetchOverviewAndSlotViz()
})

if (data?.value) {
  const [
    fetchedOverviewData,
    fetchedSlotVizData,
  ] = data.value
  handleAllData(fetchedOverviewData, fetchedSlotVizData)
}

const defaultTab = 'summary'
const activeTab = ref<string>('')
const refreshActiveTab = () => {
  switch (activeTab.value) {
    case 'blocks':
      return blocks.refresh(dashboardKey.value, defaultQueryBlocks)
    case 'deposits':
      totalClDeposits.refresh(dashboardKey.value)
      totalElDeposits.refresh(dashboardKey.value)
      clDeposits.refresh(dashboardKey.value, defaultQueryClDeposits)
      elDeposits.refresh(dashboardKey.value, defaultQueryElDeposits)
      return
    case 'rewards':
      return rewards.refresh(dashboardKey.value, defaultQueryRewards)
    case 'summary':
      return refreshSummary('last_24h', defaultQuerySummary)
    case 'withdrawals':
      totalWithdrawals.refresh(dashboardKey.value)
      withdrawals.refresh(dashboardKey.value, defaultQueryWithdrawals)
      return
  }
}
const updateTab = (tab: string) => {
  if (!isClientSide) {
    return
  }
  activeTab.value = tab
  refreshActiveTab()
}

const refreshAll = () => {
  fetchOverviewAndSlotViz()
    .then(([
      fetchedOverviewData,
      fetchedSlotVizData,
    ]) => {
      handleAllData(fetchedOverviewData, fetchedSlotVizData)
    })
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
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardControls
          :dashboard-title="dataOverview?.name"
          @dashboard-modified="refreshAll"
        />
        <DashboardValidatorOverview
          class="overview"
          :data="dataOverview"
        />
      </template>
      <DashboardSharedDashboardModal />
      <div>
        <DashboardValidatorSlotViz
          v-if="dataSlotViz"
          :data="dataSlotViz"
          @update="refreshSlotViz"
        />
      </div>
      <BcTabList
        :tabs
        :default-tab
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
        @changed-tab="updateTab"
      >
        <template #tab-panel-summary>
          <DashboardTableSummary
            :data="dataSummary"
            :paging="pagingSummary"
            :query="querySummary"
            :time-frame="timeframeSummary"
            :is-loading="isLoadingSummary"
            @update="refreshSummary"
          />
        </template>
        <template #tab-panel-rewards>
          <DashboardTableRewards
            :data="dataRewards"
            :paging="rewards.paging.value"
            :query="rewards.query.value"
            :is-loading="rewards.isLoading.value"
            @update="(query) => rewards.refresh(dashboardKey, query)"
          />
        </template>
        <template #tab-panel-blocks>
          <DashboardTableBlocks
            :data="blocks.data.value"
            :paging="blocks.paging.value"
            :query="blocks.query.value"
            :is-loading="blocks.isLoading.value"
            @update="(query) => blocks.refresh(dashboardKey, query)"
          />
        </template>
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits
              :data="elDeposits.data.value"
              :paging="elDeposits.paging.value"
              :query="elDeposits.query.value"
              :is-loading="elDeposits.isLoading.value"
              :data-total="totalElDeposits.data.value"
              :is-loading-total="totalElDeposits.isLoading.value"
              @update="(query) => elDeposits.refresh(dashboardKey, query)"
            />
            <FontAwesomeIcon
              :icon="faArrowDown"
              class="down_icon"
            />
            <DashboardTableClDeposits
              :data="clDeposits.data.value"
              :paging="clDeposits.paging.value"
              :query="clDeposits.query.value"
              :is-loading="clDeposits.isLoading.value"
              :data-total="totalClDeposits.data.value"
              :is-loading-total="totalClDeposits.isLoading.value"
              @update="(query) => elDeposits.refresh(dashboardKey, query)"
            />
          </div>
        </template>
        <template #tab-panel-withdrawals>
          <DashboardTableWithdrawals
            :data="withdrawals.data.value"
            :paging="withdrawals.paging.value"
            :query="withdrawals.query.value"
            :is-loading="withdrawals.isLoading.value"
            :data-total="totalWithdrawals.data.value"
            :is-loading-total="totalWithdrawals.isLoading.value"
            @update="(query) => elDeposits.refresh(dashboardKey, query)"
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
