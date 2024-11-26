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
import type { GuestDashboard } from '~/types/dashboard'
import {
  isGuestDashboardKey, isSharedDashboardKey,
} from '~/utils/dashboard/key'
import type { HashTabs } from '~/types/hashTabs'
import type { VDBOverviewData } from '~/types/api/validator_dashboard'
import type { SlotVizEpoch } from '~/types/api/slot_viz'

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

await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

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

const validatorDashboardStore = useValidatorDashboardStore()

const overviewData = ref<VDBOverviewData>()
const {
  fetchOverviewData,
} = useValidatorDashboardOverview()

const slotVizData = ref<SlotVizEpoch[]>()
const {
  fetchSlotVizData,
} = useValidatorSlotViz()
// fetches all data for the dashboard (overview, slot viz, active table)
function fetchAllData() {
  if (dashboardKey.value) { // valid dashboard key -> fetch all data
    return Promise.allSettled([
      fetchOverviewData(dashboardKey.value),
      fetchSlotVizData(dashboardKey.value),
    ])
  }
  if (!isLoggedIn.value) { // implies empty guest dashboard -> only fetch slot viz
    return Promise.allSettled([
      undefined,
      fetchSlotVizData(dashboardKey.value),
    ])
  }
  return Promise.allSettled([ // implies logged-in user with no dashboards -> fetch nothing
    undefined,
    undefined,
  ])
}

function handleError(e: any) {
  if (!e) {
    return
  }
  if (e.statusCode === 404) {
    // TODO: show that the dashboard does not exist
    return
  }
  throw e
}
function handleFetchedData(
  fetchedOverviewData: PromiseFulfilledResult<undefined>
    | PromiseFulfilledResult<VDBOverviewData>
    | PromiseRejectedResult,
  fetchedSlotVizData: PromiseFulfilledResult<SlotVizEpoch[]>
    | PromiseFulfilledResult<undefined>
    | PromiseRejectedResult,
) {
  if (isRejected(fetchedOverviewData)) {
    handleError(fetchedOverviewData.reason)
    return
  }
  if (isRejected(fetchedSlotVizData)) {
    handleError(fetchedSlotVizData.reason)
    return
  }
  overviewData.value = fetchedOverviewData.value
  slotVizData.value = fetchedSlotVizData.value
  if (fetchedOverviewData.value) {
    validatorDashboardStore.setByOverviewData(fetchedOverviewData.value)
  }
}
function isRejected<T>(p: PromiseSettledResult<T>): p is PromiseRejectedResult {
  return p.status === 'rejected'
}
function refreshSlotViz(groupIds: number[]) {
  fetchSlotVizData(dashboardKey.value, groupIds)
    .then((data) => {
      slotVizData.value = data
    })
    .catch((e) => {
      handleError(e)
    })
}

// init SSR data
const {
  data,
} = await useAsyncData('complete_validator_dashboard_fetch', async () => { return fetchAllData() })

if (data?.value) {
  const [
    fetchedOverviewData,
    fetchedSlotVizData,
  ] = data.value
  handleFetchedData(fetchedOverviewData, fetchedSlotVizData)
}
// updates data for all non-modal components, usually triggered by modyfing the dashboard
function updateAll() {
  fetchAllData()
    .then(([
      fetchedOverviewData,
      fetchedSlotVizData,
    ]) => {
      handleFetchedData(fetchedOverviewData, fetchedSlotVizData)
    })
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
          :dashboard-title="overviewData?.name"
          @dashboard-modified="updateAll"
        />
        <DashboardValidatorOverview
          class="overview"
          :data="overviewData"
        />
      </template>
      <DashboardSharedDashboardModal />
      <div>
        <DashboardValidatorSlotViz
          :data="slotVizData"
          @update="refreshSlotViz"
        />
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
            <FontAwesomeIcon
              :icon="faArrowDown"
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
