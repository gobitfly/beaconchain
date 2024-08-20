<script setup lang="ts">
import {
  faBolt,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faNetworkWired,
} from '@fortawesome/pro-solid-svg-icons'
import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import {
  BcDialogConfirm, NotificationsNetworkTable,
} from '#components'
import type { HashTabs } from '~/types/hashTabs'

useDashboardKeyProvider('notifications')
const { refreshDashboards } = useUserDashboardStore()
const { isLoggedIn } = useUserStore()
const dialog = useDialog()
const { t: $t } = useTranslation()

await useAsyncData('user_dashboards', () => refreshDashboards(), { watch: [ isLoggedIn ] })

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const manageNotificationsModalVisisble = ref(false)

const tabs: HashTabs = {
  clients: {
    disabled: !showInDevelopment,
    icon: faBolt,
    index: '2',
    placeholder: 'Clients coming soon!',
    title: $t('notifications.tabs.clients'),
  },
  dashboards: {
    icon: faGaugeSimpleMax,
    index: '0',
    title: $t('notifications.tabs.dashboards'),
  },
  machines: {
    disabled: !showInDevelopment,
    icon: faMonitorWaveform,
    index: '1',
    placeholder: 'Machines coming soon!',
    title: $t('notifications.tabs.machines'),
  },
  network: {
    component: NotificationsNetworkTable,
    disabled: !showInDevelopment,
    icon: faNetworkWired,
    index: '4',
    title: $t('notifications.tabs.network'),
  },
  rocketpool: {
    disabled: !showInDevelopment,
    index: '3',
    placeholder: 'Rocketpool coming soon!',
    title: $t('notifications.tabs.rocketpool'),
  },
}

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
          TODO: Overview
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
        :tabs default-tab="dashboards"
        :use-route-hash="true"
        class="notifications-tab-view"
        panels-class="notifications-tab-panels"
      >
        <template #tab-header-icon-rocketpool>
          <IconRocketPool />
        </template>
        <template #tab-panel-dashboards>
          <NotificationsDashboardsTable
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
