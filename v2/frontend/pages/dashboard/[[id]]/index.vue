<script setup lang="ts">
import {
  faChartLineUp,
  faCube,
  faCubes,
  faFire,
  faWallet,
  faMoneyBill,
  faShare,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { DashboardCreationController, BcDialogConfirm } from '#components'
import type { CookieDashboard, DashboardKey } from '~/types/dashboard'
import type { MenuBarEntry } from '~/types/menuBar'

const { isLoggedIn } = useUserStore()

const { dashboardKey, setDashboardKey, isPublic } = useDashboardKeyProvider()
const { refreshDashboards, updateHash, dashboards } = useUserDashboardStore()

const dialog = useDialog()
const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const router = useRouter()
const { width } = useWindowSize()

const manageButtons = computed<MenuBarEntry[] | undefined>(() => {
  if (width.value < 520 && isLoggedIn.value && !isPublic.value) {
    return [
      {
        label: 'Manage',
        dropdown: true,
        items: [
          {
            label: $t('dashboard.validator.manage_groups'),
            command: () => { manageGroupsModalVisisble.value = true }
          },
          {
            label: $t('dashboard.validator.manage_validators'),
            command: () => { manageValidatorsModalVisisble.value = true }
          }
        ]
      }
    ]
  }

  return [
    {
      dropdown: false,
      label: $t('dashboard.validator.manage_groups'),
      command: () => { manageGroupsModalVisisble.value = true }
    },
    {
      dropdown: false,
      label: $t('dashboard.validator.manage_validators'),
      command: () => { manageValidatorsModalVisisble.value = true }
    }
  ]
})

const { refreshOverview } = useValidatorDashboardOverviewStore()
await Promise.all([
  useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [isLoggedIn] }),
  useAsyncData('validator_overview', () => refreshOverview(dashboardKey.value), { watch: [dashboardKey] })
])

const manageValidatorsModalVisisble = ref(false)
const manageGroupsModalVisisble = ref(false)

const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreationDialog () {
  dashboardCreationControllerModal.value?.show()
}

const share = () => {
  alert('Not implemented yet')
}

const remove = () => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.deletion.title')
    },
    onClose: response => response?.data && removeDashboard(dashboardKey.value),
    data: {
      question: $t('dashboard.deletion.text', { dashboard: 'dashboardName.value' }) // TODO: Fix
    }
  })
}

const removeDashboard = async (key: DashboardKey) => {
  await fetch(API_PATH.DASHBOARD_DELETE_VALIDATOR, { body: { key } }, { dashboardKey: key })

  await refreshDashboards()

  // forward user to another dashboard (if possible)
  if ((dashboards.value?.validator_dashboards?.length ?? 0) > 0) {
    router.push(`/dashboard/${dashboards.value?.validator_dashboards[0].id}`)
    return
  }

  if ((dashboards.value?.account_dashboards?.length ?? 0) > 0) {
    router.push(`/account-dashboard/${dashboards.value?.account_dashboards[0].id}`)
    return
  }

  // no other dashboard available, forward to creation screen
  router.push('/dashboard')
}

onMounted(() => {
  if (dashboardKey.value === '') {
    // we don't have a key and no validator dashboard: show the create panel
    if (dashboards.value?.validator_dashboards?.length) {
      // if we have a validator dashboard but none selected: select the first
      const cd = dashboards.value.validator_dashboards[0] as CookieDashboard
      setDashboardKey(cd.hash ?? cd.id.toString())
    }
  }
})

watch(dashboardKey, (newKey, oldKey) => {
  if (!isLoggedIn.value) {
    // We update the key for our public dashboard
    const cd = dashboards.value?.validator_dashboards?.[0] as CookieDashboard
    // If the old key does not match the dashboards key then it probabbly means we opened a different pub. dashboard as a link
    if (cd && (!cd.hash || (cd.hash ?? '') === (oldKey ?? ''))) {
      updateHash('validator', newKey)
    }
  }
})
</script>

<template>
  <div v-if="!dashboardKey && !dashboards?.validator_dashboards?.length">
    <BcPageWrapper>
      <DashboardCreationController
        class="panel-controller"
        :display-type="'panel'"
        :initially-visislbe="true"
      />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" />
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" />
    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-type="'modal'"
    />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreationDialog()" />
        <DashboardValidatorOverview class="overview" />
      </template>
      <div class="header-row" :class="{'single-element':!(isLoggedIn && !isPublic)}">
        <div v-if="isLoggedIn && !isPublic" class="action-button-container">
          <Button class="share-button" @click="share()">
            {{ $t('dashboard.share') }}<FontAwesomeIcon :icon="faShare" />
          </Button>
          <Button class="p-button-icon-only" @click="remove()">
            <FontAwesomeIcon :icon="faTrash" />
          </Button>
        </div>
        <Menubar v-if="manageButtons" :model="manageButtons" breakpoint="0px" class="right-aligned-submenu">
          <template #item="{ item }">
            <span class="button-content pointer">
              <span class="text">{{ item.label }}</span>
              <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
            </span>
          </template>
        </Menubar>
      </div>
      <div>
        <DashboardValidatorSlotViz />
      </div>
      <TabView lazy>
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
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.heatmap')" :icon="faFire" />
          </template>
          Heatmap coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.deposits')" :icon="faWallet" />
          </template>
          Deposits coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.withdrawals')" :icon="faMoneyBill" />
          </template>
          Withdrawals coming soon!
        </TabPanel>
      </TabView>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.header-row {
  height: 30px;
  display: flex;
  justify-content: space-between;
  gap: var(--padding);
  margin-bottom: var(--padding);

  &.single-element {
    justify-content: flex-end;
  }

  .action-button-container{
    display: flex;
    gap: var(--padding);

    .share-button{
      display: flex;
      gap: var(--padding-small);
    }
  }

  :deep(.p-menubar .p-menubar-root-list) {
    >.p-menuitem{
      height: 30px;
      color: var(--text-color-inverted);
      background: var(--button-color-active);
      font-weight: var(--standard_text_medium_font_weight);
      border-color: var(--button-color-active);

      >.p-menuitem-content {
        padding: 0;

        >.button-content{
          display: flex;
          align-items: center;
          gap: 7px;
          padding: 7px 17px;

          .pointer {
            cursor: pointer;
          }
        }
      }

      >.p-submenu-list {
        font-weight: var(--standard_text_font_weight);

        >.p-menuitem .button-content{
          gap: 0;
          padding: 0;

          .text {
            @include utils.truncate-text;
          }
        }
      }

      &:not(.p-highlight):not(.p-disabled) > .p-menuitem-content:hover {
        background: var(--button-color-hover);
      }
    }
  }
}

.panel-controller {
  display: flex;
  justify-content: center;
  margin-top: 60px;
  margin-bottom: 60px;
  overflow: hidden;
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
</style>
