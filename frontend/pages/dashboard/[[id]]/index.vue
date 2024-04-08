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
import { DashboardCreationController, BcDialogConfirm, IconMore } from '#components'
import type { DashboardCreationDisplayType } from '~/types/dashboard/creation'
import type { DashboardKey } from '~/types/dashboard'
import type { MenuBarEntry } from '~/types/menuBar'

const route = useRoute()
const dialog = useDialog()
const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const router = useRouter()
const { width } = useWindowSize()

const key = computed<DashboardKey>(() => {
  if (Array.isArray(route.params.id)) {
    return route.params.id.join(',')
  }

  const idAsNumber = parseInt(route.params.id)
  if (isNaN(idAsNumber)) {
    return route.params.id
  }

  return idAsNumber
})

const moreButtons = computed<MenuBarEntry[] | undefined>(() => {
  if (width.value < 525) {
    return [
      {
        label: '',
        dropdown: false,
        class: 'icon-only',
        component: IconMore,
        items: [
          {
            label: $t('dashboard.share_dashboard'),
            command: () => { share() }
          },
          {
            label: $t('dashboard.delete_dashboard'),
            command: () => { remove() }
          }
        ]
      }
    ]
  }

  return undefined
})

const manageButtons = computed<MenuBarEntry[] | undefined>(() => {
  if (width.value < 850) {
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

  return undefined
})

const { dashboards, refreshDashboards, getValidatorDashboardName } = useUserDashboardStore()
const { refreshOverview } = useValidatorDashboardOverviewStore()
await Promise.all([
  useAsyncData('user_dashboards', () => refreshDashboards()),
  useAsyncData('validator_overview', () => refreshOverview(key.value), { watch: [key] })
])

const dashboardName = computed(() => getValidatorDashboardName(key.value))

const manageValidatorsModalVisisble = ref(false)
const manageGroupsModalVisisble = ref(false)

const dashboardCreationControllerPanel = ref<typeof DashboardCreationController>()
const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreation (type: DashboardCreationDisplayType) {
  if (type === 'panel') {
    dashboardCreationControllerPanel.value?.show()
  } else {
    dashboardCreationControllerModal.value?.show()
  }
}

const share = () => {
  alert('Not implemented yet')
}

const remove = () => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.deletion.title')
    },
    onClose: response => response?.data && removeDashboard(key.value),
    data: {
      question: $t('dashboard.deletion.text', { dashboard: dashboardName.value })
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
  // TODO: Implement check if user does not have a single dashboard instead of the key check once information is available
  if (key.value === '') {
    showDashboardCreation('panel')
  }
})
</script>

<template>
  <div v-if="key === ''">
    <BcPageWrapper>
      <DashboardCreationController
        ref="dashboardCreationControllerPanel"
        class="panel-controller"
        :display-type="'panel'"
      />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" :dashboard-key="key" />
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" :dashboard-key="key" />
    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-type="'modal'"
    />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreation('modal')" />
        <DashboardValidatorOverview class="overview" :dashboard-key="key" />
      </template>
      <div class="header-row">
        <div class="name-container">
          <div class="h1 name">
            {{ dashboardName }}
          </div>
          <Menubar v-if="moreButtons" :model="moreButtons" breakpoint="0px">
            <template #item="{ item }">
              <span class="button-content more-button pointer">
                <div v-if="item.component">
                  <component :is="item.component" />
                </div>
                <div v-else>
                  <span class="text">{{ item.label }}</span>
                  <IconChevron v-if="item.dropdown" direction="bottom" />
                </div>
              </span>
            </template>
          </Menubar>
          <div v-else class="button-container">
            <Button class="share-button" @click="share()">
              {{ $t('dashboard.share') }}<FontAwesomeIcon :icon="faShare" />
            </Button>
            <Button class="p-button-icon-only" @click="remove()">
              <FontAwesomeIcon :icon="faTrash" />
            </Button>
          </div>
        </div>
        <Menubar v-if="manageButtons" :model="manageButtons" breakpoint="0px" class="right-aligned-submenu">
          <template #item="{ item }">
            <span class="button-content pointer">
              <span class="text">{{ item.label }}</span>
              <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
            </span>
          </template>
        </Menubar>
        <div v-else class="manage-buttons-container">
          <Button :label="$t('dashboard.validator.manage_groups')" @click="manageGroupsModalVisisble = true" />
          <Button :label="$t('dashboard.validator.manage_validators')" @click="manageValidatorsModalVisisble = true" />
        </div>
      </div>
      <div>
        <DashboardValidatorSlotViz :dashboard-key="key" />
      </div>
      <TabView lazy>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.summary')" :icon="faChartLineUp" />
          </template>
          <DashboardTableSummary :dashboard-key="key" />
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.rewards')" :icon="faCubes" />
          </template>
          Rewards coming soon!
        </TabPanel>
        <TabPanel>
          <template #header>
            <BcTabHeader :header="$t('dashboard.validator.tabs.blocks')" :icon="faCube" />
          </template>
          Blocks coming soon!
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

  .name-container{
    display: flex;
    gap: var(--padding-large);

    max-width: 900px;
    @media (max-width: 1260px) {
      max-width: calc(100% - (var(--padding)*3) - var(--padding-large) - 330px);
    }
    @media (max-width: 849px) {
      max-width: calc(100% - (var(--padding)*2) - var(--padding-large) - 110px);
    }

    .name {
      margin-top: 0;
      @include utils.truncate-text;
    }

    .button-container{
      display: flex;
      gap: var(--padding);

      .share-button{
        display: flex;
        gap: var(--padding-small);
      }
    }
  }

  .manage-buttons-container{
    display: flex;
    justify-content: flex-end;
    gap: var(--padding);
  }

  :deep(.p-menubar .p-menubar-root-list) {
    >.p-menuitem{
      color: var(--text-color-inverted);
      background: var(--button-color-active);
      font-weight: var(--standard_text_medium_font_weight);
      border-color: var(--button-color-active);

      >.p-menuitem-content {
        padding: 0;

        >.button-content{
          display: flex;

          &:not(.more-button){
            align-items: center;
            gap: 7px;
            padding: 7px 17px;
          }

          &.more-button{
            padding: 6px 13px 3px 13px;
          }

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
