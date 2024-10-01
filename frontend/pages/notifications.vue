<script setup lang="ts">
import {
  faBolt,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faNetworkWired,
} from '@fortawesome/pro-solid-svg-icons'
import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import { BcDialogConfirm } from '#components'
import type { HashTabs } from '~/types/hashTabs'

// TODO: get rid of this provider logic -> necessary for `<DashboardHeader`
useDashboardKeyProvider('notifications')

const { isLoggedIn } = useUserStore()
const dialog = useDialog()
const { t: $t } = useTranslation()

const manageNotificationsModalVisisble = ref(false)
const tabKey = {
  clients: 'clients',
  dashboards: 'dashboards',
  machines: 'machines',
  network: 'network',
  rocketpool: 'rocketpool',
}
const tabs: HashTabs = [
  {
    icon: faGaugeSimpleMax,
    key: tabKey.dashboards,
    title: $t('notifications.tabs.dashboards'),
  },
  {
    icon: faMonitorWaveform,
    key: tabKey.machines,
    title: $t('notifications.tabs.machines'),
  },
  {
    icon: faBolt,
    key: tabKey.clients,
    title: $t('notifications.tabs.clients'),
  },
  {
    key: tabKey.rocketpool,
    title: $t('notifications.tabs.rocketpool'),
  },
  {
    icon: faNetworkWired,
    key: tabKey.network,
    title: $t('notifications.tabs.network'),
  },
]

const getSlotName = (key: string) => `tab-panel-${key}`

useBcSeo('notifications.title')

const openManageNotifications = () => {
  if (!isLoggedIn.value) {
    dialog.open(BcDialogConfirm, {
      data: { question: $t('notifications.login_question') },
      onClose: async (response: DynamicDialogCloseOptions) => {
        if (response?.data) {
          await navigateTo('/login')
        }
      },
      props: { header: $t('notifications.title') },
    })
  }
  else {
    manageNotificationsModalVisisble.value = true
  }
}
</script>

<template>
  <div>
    <BcPageWrapper>
      <template #top>
        <DashboardHeader :dashboard-title="$t('notifications.title')" />
        <div class="overview">
          <NotificationsOverview
            @open-dialog="openManageNotifications"
          />
        </div>
      </template>
      <NotificationsManagementModal
        v-model="manageNotificationsModalVisisble"
      />
      <div class="button-row">
        <Button
          :label="$t('notifications.manage')"
          @click="openManageNotifications"
        />
      </div>
      <BcTabList
        :tabs
        default-tab="dashboards"
        :use-route-hash="true"
        class="notifications-tab-view"
        panels-class="notifications-tab-panels"
      >
        <template #[getSlotName(tabKey.dashboards)]>
          <NotificationsDashboardsTable
            @open-dialog="openManageNotifications"
          />
        </template>
        <template #[getSlotName(tabKey.clients)]>
          <NotificationsClientsTable
            @open-dialog="openManageNotifications"
          />
        </template>
        <template #tab-header-icon-rocketpool>
          <IconRocketPool />
        </template>
        <template #[getSlotName(tabKey.rocketpool)]>
          <NotificationsRocketPoolTable
            @open-dialog="openManageNotifications"
          />
        </template>
        <template #[getSlotName(tabKey.network)]>
          <NotificationsNetworkTable
            @open-dialog="openManageNotifications"
          />
        </template>
        <template #[getSlotName(tabKey.machines)]>
          <NotificationsMachinesTable
            @open-dialog="openManageNotifications"
          />
        </template>
      </BcTabList>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>
.overview {
  margin-bottom: var(--padding-large);
}

.notifications-tab-view {
  margin-top: var(--padding-large);
  :deep(.notifications-tab-panels) {
    min-height: 699px;
  }
}

.button-row {
  display: flex;
  justify-content: flex-end;
}
</style>
