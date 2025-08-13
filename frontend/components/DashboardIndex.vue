<script setup lang="ts">
// import type { GuestDashboard } from '~/types/dashboard'
// import {
//   isGuestDashboardKey, isSharedDashboardKey,
// } from '~/utils/dashboard/key'
import type { HashTab } from '~/components/bc/tab/BcTabList.vue'
import type { ValidatorDashboard } from '~/types/api/dashboard'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import type {
  VDBOverviewData,
  VDBOverviewGroup,
} from '~/types/api/validator_dashboard'

// import type { TableQueryParams } from '~/types/datatable'

// const route = useRoute()

// onBeforeMount(() => {
// const { validators } = route.query
// if (validators && typeof validators === 'string') {
//   const filteredValidators = validators
//     .split(',')
//     .filter(
//       validatorId => isNumber(validatorId) || isPublicKey(validatorId),
//     )
//     .join(',')
//   return navigateTo(`/dashboard/${encodeBase64Url(filteredValidators)}`, { replace: true })
// }
// })

// const {
//   isLoggedIn,
// } = useUserStore()
const { t: $t } = useTranslation()

const tabs: HashTab[] = [
  {
    // component: DashboardTableSummary,
    icon: 'chart-line-up',
    key: 'summary',
    title: $t('dashboard.validator.tabs.summary'),
  },
  {
    // component: DashboardTableRewards,
    icon: 'cubes',
    key: 'rewards',
    title: $t('dashboard.validator.tabs.rewards'),
  },
  {
    // component: DashboardTableBlocks,d
    icon: 'cube',
    key: 'blocks',
    title: $t('dashboard.validator.tabs.blocks'),
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

// const {
// dashboardKey,
// setDashboardKey,
// } = useDashboardKeyProvider('validator')

// const userDashboardStore = useUserDashboardStore()
// const {
// getDashboardLabel,
// refreshDashboards,
// updateGuestDashboardKey,
// } = userDashboardStore

// const {
//   cookieDashboards,
//   dashboards,
// } = storeToRefs(userDashboardStore)

// const seoTitle = computed(() => {
// return getDashboardLabel(dashboardKey.value, 'validator')
// })

// useBcSeo(seoTitle, true)

// const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
// const {
// hasValidators,
//   overview,
// } = storeToRefs(validatorDashboardOverviewStore)
// const {
//   refreshOverview,
// } = validatorDashboardOverviewStore

// const {
//   getProducts,
// } = useProductsStore()

// await useAsyncData('get_products', () => getProducts())

// await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

// const { error: validatorOverviewError } = await useAsyncData(
//   'validator_overview',
//   () => refreshOverview(dashboardKey.value),
//   { watch: [ dashboardKey ] },
// )
// when we run into an error loading a dashboard keep it here to prevent an infinity loop
// const errorDashboardKeys: string[] = []
// const setDashboardKeyIfNoError = (key: string) => {
//   if (!errorDashboardKeys.includes(key)) {
//     setDashboardKey(key)
//   }
// }

// const dashboardCreationControllerModal
//   = ref<typeof DashboardCreationController>()
// function showDashboardCreationDialog() {
//   dashboardCreationControllerModal.value?.show()
// }

// watch(
//   [
//     dashboardKey,
//     isLoggedIn,
//   ],
//   ([
//     newKey,
//     newLoggedIn,
//   ], [ oldKey ]) => {
//     if (!newLoggedIn || !newKey) {
//       // Some checks if we need to update the dashboard key or the guest dashboard
//       let gd = dashboards.value?.validator_dashboards?.[0] as GuestDashboard
//       const isGuest = isGuestDashboardKey(newKey)
//       const isShared = isSharedDashboardKey(newKey)
//       if (isShared) {
//         return
//       }
//       if (newLoggedIn) {
//         // if we are logged in and have no dashboard key we only want to switch
//         //  to the first dashboard if it is a private one
//         if (gd && gd.key === undefined) {
//           setDashboardKeyIfNoError(gd.id.toString())
//         }
//       }
//       else if (
//         !newLoggedIn
//         && gd
//         && isGuest
//         && (!gd.key || (gd.key ?? '') === (oldKey ?? ''))
//       ) {
//         // we got a new guest dashboard key but the old key matches the
//         // stored dashboard - so we update the stored dashboard
//         if (!errorDashboardKeys.includes(newKey)) {
//           updateGuestDashboardKey('validator', newKey)
//         }
//         setDashboardKeyIfNoError(newKey ?? '')
//       }
//       else if (!newKey || !isGuest) {
//         // trying to view a private dashboad but not logged in
//         gd = cookieDashboards.value
//           ?.validator_dashboards?.[0] as GuestDashboard
//         setDashboardKeyIfNoError(gd?.key ?? '')
//       }
//     }
//   },
//   { immediate: true },
// )

// const activeTab = computed(() => route.hash)

// const refreshActiveTab = () => {
//   if (!hasValidators.value) return

//   switch (activeTab.value) {
//     case '#consolidations':
//       refreshElConsolidationsData()
//       refreshClConsolidationsData()
//       break
//     case '#deposits':
//       refreshElDepositsData()
//       refreshClDepositsData()
//       break
//     case '#withdrawals':
//       refreshElWithdrawalsData()
//       refreshClWithdrawalsData()
//       break
//   }
// }

// watch([
//   activeTab,
//   overview,
// ],
// () => {
//   refreshActiveTab()
// },
// {
//   immediate: true,
// },
// )

defineProps<{
  overview: null | VDBOverviewData,
  slotVizEpochs: null | SlotVizEpoch[],
  validatorDashboards: null | ValidatorDashboard[],
}>()
const visible = ref(false)
const onShowCreation = () => {
  visible.value = true
}

const isVisibleManagementModal = ref(false)
const onAddValidator = () => {
  isVisibleManagementModal.value = true
}

const emit = defineEmits<{
  (e: 'change-validators', validators: string[]): void,
  (e: 'change-groups', value: VDBOverviewGroup[]): void,
}>()

// const tab = ref('summary')

// const onTabChange = (value: string) => {
//   tab.value = value
// }
const { key } = useDashboard()
</script>

<template>
  <!-- <NuxtLayout v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <DashboardCreationController
      class="panel-controller"
      :display-mode="'panel'"
      :initially-visible="true"
    />
  </NuxtLayout> -->
  <NuxtLayout name="default">
    <BcDialog v-model="visible">
      <LazyDashboardCreationController
        v-if="visible"
        ref="dashboardCreationControllerModal"
        class="modal-controller"
      />
    </BcDialog>
    <template #banner>
      <LazyBcNotificationBanner
        v-if=" overview?.is_above_effective_balance_limit"
        :title="$t('dashboard.subsciprion_limit_reached_title')"
      >
        <LazyBcTranslation
          v-if=" overview?.is_above_effective_balance_limit"
          keypath="dashboard.subsciprion_limit_reached.template"
          linkpath="dashboard.subsciprion_limit_reached._link"
          to="/pricing"
        />
      </LazyBcNotificationBanner>
    </template>
    <DashboardHeader @show-creation="onShowCreation()" />
    <DashboardControls
      v-model:is-visible-management-modal="isVisibleManagementModal"
      :validator-dashboards
      :dashboard-title="overview?.name ?? ''"
      @change-validators="emit('change-validators', $event)"
      @change-groups="emit('change-groups', $event)"
    />
    <DashboardValidatorOverview
      :overview
      class="overview"
    />
    <DashboardSharedDashboardModal />
    <DashboardSlotViz
      :slot-viz-epochs="slotVizEpochs || []"
    />
    <BcTabList
      :tabs
      query-parameter-key="tab"
      default-tab="summary"
      class="dashboard-tab-view"
      panels-class="dashboard-tab-panels"
    >
      <template #tab-panel-summary="{ isActive }">
        <LazyDashboardTableSummary
          v-if="key && isActive"
        />
      </template>
      <template #tab-panel-rewards="{ isActive }">
        <LazyDashboardTableRewards
          v-if="key && isActive"
        />
      </template>
      <template #tab-panel-blocks="{ isActive }">
        <LazyDashboardTableBlocks
          v-if="key && isActive"
        />
      </template>
      <template #tab-panel-deposits="{ isActive }">
        <div v-if="key && isActive">
          <LazyDashboardTableElDeposits />
          <LazyBcIcon
            name="arrow-down"
            class="down_icon"
          />
          <LazyDashboardTableClDeposits />
        </div>
      </template>
      <template #tab-panel-withdrawals="{ isActive }">
        <div v-if="key && isActive">
          <LazyDashboardTableElWithdrawals />
          <LazyBcIcon
            name="arrow-down"
            class="down_icon"
          />
          <LazyDashboardTableClWithdrawals />
        </div>
      </template>
      <template #tab-panel-consolidations="{ isActive }">
        <div v-if="key && isActive">
          <LazyDashboardTableElConsolidations />
          <LazyBcIcon
            name="arrow-down"
            class="down_icon"
          />
          <LazyDashboardTableClConsolidations />
        </div>
      </template>
      <template #empty>
        <LazyDashboardTableAddValidator
          v-if="!key"
          @add-validator="onAddValidator"
        />
      </template>
    </BcTabList>
  </NuxtLayout>
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
