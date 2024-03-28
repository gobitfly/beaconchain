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
import type { DashboardCreationController } from '#components'
import type { DashboardCreationDisplayType } from '~/types/dashboard/creation'
import type { DashboardKey } from '~/types/dashboard'

const route = useRoute()

const key = computed<DashboardKey>(() => {
  if (Array.isArray(route.params.id)) {
    return route.params.id.join(',')
  }
  return route.params.id
})

const { refreshDashboards } = useUserDashboardStore()
const { refreshOverview } = useValidatorDashboardOverviewStore()
await Promise.all([
  useAsyncData('user_dashboards', () => refreshDashboards()),
  useAsyncData('validator_overview', () => refreshOverview(key.value), { watch: [key] })
])

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

const onShare = () => {
  alert('Not implemented yet')
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
            Validators
          </div>
          <div class="button-container">
            <Button class="share-button" @click="onShare()">
              {{ $t('dashboard.validator.share') }}<FontAwesomeIcon :icon="faShare" />
            </Button>
            <Button class="p-button-icon-only">
              <FontAwesomeIcon :icon="faTrash" />
            </Button>
          </div>
        </div>
        <div class="manage-buttons-container">
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
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;

  .name-container{
    display: flex;
    align-items: center;
    gap: var(--padding-large);

    .name {
      margin-top: 0;
    }

    .button-container{
      display: flex;
      gap: var(--padding);

      .share-button{
        display: flex;
        align-items: center;
        gap: var(--padding-small);
      }
    }
  }

  .manage-buttons-container{
    display: flex;
    justify-content: flex-end;
    gap: var(--padding);
    margin-bottom: var(--padding);
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
