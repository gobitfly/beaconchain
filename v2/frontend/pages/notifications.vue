<script setup lang="ts">
import {
  faBolt,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faNetworkWired,
} from '@fortawesome/pro-solid-svg-icons'
import { BcDialogConfirm } from '#components'
import type { HashTabs } from '~/types/hashTabs'

useDashboardKeyProvider('notifications')
const { refreshDashboards } = useUserDashboardStore()
const { isLoggedIn } = useUserStore()
const dialog = useDialog()
const { t: $t } = useTranslation()

await useAsyncData('user_dashboards', () => refreshDashboards(), {
  watch: [isLoggedIn],
})

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const manageNotificationsModalVisisble = ref(false)

const tabs: HashTabs = {
  clients: {
    disabled: !showInDevelopment,
    index: 2,
  },
  dashboards: {
    index: 0,
  },
  machines: {
    disabled: !showInDevelopment,
    index: 1,
  },
  network: {
    disabled: !showInDevelopment,
    index: 4,
  },
  rocketpool: {
    disabled: !showInDevelopment,
    index: 3,
  },
}

const { activeIndex, setActiveIndex } = useHashTabs(tabs)

useBcSeo('notifications.title')

const openManageNotifications = () => {
  if (!isLoggedIn.value) {
    dialog.open(BcDialogConfirm, {
      data: {
        question: $t('notifications.login_question'),
      },
      onClose: async (response) => {
        if (response?.data) {
          await navigateTo('/login')
        }
      },
      props: {
        header: $t('notifications.title'),
      },
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
      <TabView
        lazy
        class="notifications-tab-view"
        :active-index="activeIndex"
        @update:active-index="setActiveIndex"
      >
        <TabPanel>
          <template #header>
            <BcTabHeader
              :header="$t('notifications.tabs.dashboards')"
              :icon="faGaugeSimpleMax"
            />
          </template>
          <NotificationsDashboardsTable
            @open-dialog="openManageNotifications"
          />
        </TabPanel>
        <TabPanel :disabled="tabs.machines.disabled">
          <template #header>
            <BcTabHeader
              :header="$t('notifications.tabs.machines')"
              :icon="faMonitorWaveform"
            />
          </template>
          Machines coming soon!
        </TabPanel>
        <TabPanel :disabled="tabs.clients.disabled">
          <template #header>
            <BcTabHeader
              :header="$t('notifications.tabs.clients')"
              :icon="faBolt"
            />
          </template>
          Clients coming soon!
        </TabPanel>
        <TabPanel :disabled="tabs.rocketpool.disabled">
          <template #header>
            <BcTabHeader :header="$t('notifications.tabs.rocketpool')">
              <template #icon>
                <IconRocketPool />
              </template>
            </BcTabHeader>
          </template>
          Rocketpool coming soon!
        </TabPanel>
        <TabPanel :disabled="tabs.network.disabled">
          <template #header>
            <BcTabHeader
              :header="$t('notifications.tabs.network')"
              :icon="faNetworkWired"
            />
          </template>
          Network coming soon!
        </TabPanel>
      </TabView>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>
:global(.notifications-tab-view > .p-tabview-panels) {
  min-height: 699px;
}

.overview {
  margin-bottom: var(--padding-large);
}

.p-tabview {
  margin-top: var(--padding-large);
}

.button-row {
  display: flex;
  justify-content: flex-end;
}
</style>
