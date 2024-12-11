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
  type TableProps,
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
import type {
  ApiPagingResponse, Paging,
} from '~/types/api/common'
import type { SummaryTimeFrame } from '~/types/dashboard/summary'

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

const dataSummary = ref<VDBSummaryTableRow[]>()
const pagingSummary = ref<Paging>()
const isLoadingSummary = ref(false)
const defaultQuerySummary: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'efficiency:desc',
}
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

const propsSummary = computed(() => ({
  data: dataSummary.value,
  isLoading: isLoadingSummary.value,
  paging: pagingSummary.value,
  query: querySummary.value,
  timeFrame: timeframeSummary.value,
}))

// function that creates appropriate table refs and refresh functions
const createTableHandler = <T extends object>(
  fetchFunc: (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => Promise<ApiPagingResponse<T> | undefined>,
  defaultQuery: TableQueryParams,
) => {
  const data = ref<T[]>()
  const paging = ref<Paging>()
  const isLoading = ref(false)
  const query = ref<TableQueryParams>(defaultQuery)

  const refresh = (inputQuery: TableQueryParams) => {
    isLoading.value = true
    query.value = inputQuery
    fetchFunc(dashboardKey.value, inputQuery)
      .then((fetchedResult) => {
        data.value = fetchedResult?.data
        paging.value = fetchedResult?.paging
      })
      .catch((e) => {
        handleError(e)
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  const props = computed<TableProps<T>>(() => ({
    data: data.value,
    isLoading: isLoading.value,
    paging: paging.value,
    query: query.value,
  }))

  return {
    props,
    refresh,
  }
}

const defaultQueryRewards: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'epoch:desc',
}
const {
  props: rawPropsRewards,
  refresh: refreshRewards,
} = createTableHandler(service.fetchRewards, defaultQueryRewards)

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

// Computed property for rewards data with a potential "future row"
const rewardsData = computed(() => {
  const data = rawPropsRewards.value.data
  if (!data || data.length === 0) {
    return undefined
  }

  const isFirstPage = !rawPropsRewards.value.paging?.prev_cursor
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

const propsRewards = computed(() => ({
  data: rewardsData.value,
  isLoading: rawPropsRewards.value.isLoading,
  paging: rawPropsRewards.value.paging,
  query: rawPropsRewards.value.query,
}))

const defaultQueryBlocks: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}
const {
  props: propsBlocks,
  refresh: refreshBlocks,
} = createTableHandler(service.fetchBlocks, defaultQueryBlocks)

const createDataHandlers = <T extends object>(
  fetchFunc: (
    dashboardKey: DashboardKey
  ) => Promise<T | undefined>,
) => {
  const data = ref<T>()
  const isLoading = ref(false)
  const refresh = async () => {
    isLoading.value = true
    fetchFunc(dashboardKey.value)
      .then((fetchedData) => {
        data.value = fetchedData
      })
      .catch((e) => {
        handleError(e)
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  const props = computed(() => ({
    data: data.value,
    isLoading: isLoading.value,
  }))
  return {
    props,
    refresh,
  }
}

const defaultQueryClDeposits: TableQueryParams = {
  limit: 5,
}
const {
  props: propsClDeposits,
  refresh: refreshClDeposits,
} = createTableHandler(service.fetchClDeposits, defaultQueryClDeposits)
const {
  props: propsTotalClDeposits,
  refresh: refreshTotalClDeposits,
} = createDataHandlers(service.fetchTotalClDeposits)
const defaultQueryElDeposits: TableQueryParams = {
  limit: 5,
}
const {
  props: propsElDeposits,
  refresh: refreshElDeposits,
} = createTableHandler(service.fetchElDeposits, defaultQueryElDeposits)
const {
  props: propsTotalElDeposits,
  refresh: refreshTotalElDeposits,
} = createDataHandlers(service.fetchTotalElDeposits)
const defaultQueryWithdrawals: TableQueryParams = {
  limit: defaultPageSize,
  sort: 'slot:desc',
}
const {
  props: propsWithdrawals,
  refresh: refreshWithdrawals,
} = createTableHandler(service.fetchWithdrawals, defaultQueryWithdrawals)
const {
  props: propsTotalWithdrawals,
  refresh: refreshTotalWithdrawals,
} = createDataHandlers(service.fetchTotalWithdrawals)

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

const refreshActiveTab = () => {
  switch (activeTab.value) {
    case 'blocks':
      return refreshBlocks(defaultQueryBlocks)
    case 'deposits':
      refreshTotalClDeposits()
      refreshTotalElDeposits()
      refreshClDeposits(defaultQueryClDeposits)
      refreshElDeposits(defaultQueryElDeposits)
      return
    case 'rewards':
      return refreshRewards(defaultQueryRewards)
    case 'summary':
      return refreshSummary('last_24h', defaultQuerySummary)
    case 'withdrawals':
      refreshWithdrawals(defaultQueryWithdrawals)
      refreshTotalWithdrawals()
      return
  }
}

const activeTab = ref<string>('summary')
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
          :data="dataSlotViz"
          @update="refreshSlotViz"
        />
      </div>
      <BcTabList
        :tabs
        default-tab="summary"
        :use-route-hash="true"
        class="dashboard-tab-view"
        panels-class="dashboard-tab-panels"
        @changed-tab="updateTab"
      >
        <template #tab-panel-summary>
          <DashboardTableSummary
            v-bind="propsSummary"
            @update="refreshSummary"
          />
        </template>
        <template #tab-panel-rewards>
          <DashboardTableRewards
            v-bind="propsRewards"
            @update="refreshRewards"
          />
        </template>
        <template #tab-panel-blocks>
          <DashboardTableBlocks
            v-bind="propsBlocks"
            @update="refreshBlocks"
          />
        </template>
        <template #tab-panel-deposits>
          <div class="deposits">
            <DashboardTableElDeposits
              :table-props="propsElDeposits"
              :total-props="propsTotalElDeposits"
              @update="refreshElDeposits"
            />
            <FontAwesomeIcon
              :icon="faArrowDown"
              class="down_icon"
            />
            <DashboardTableClDeposits
              :table-props="propsClDeposits"
              :total-props="propsTotalClDeposits"
              @update="refreshClDeposits"
            />
          </div>
        </template>
        <template #tab-panel-withdrawals>
          <DashboardTableWithdrawals
            :table-props="propsWithdrawals"
            :total-props="propsTotalWithdrawals"
            @update="refreshWithdrawals"
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
