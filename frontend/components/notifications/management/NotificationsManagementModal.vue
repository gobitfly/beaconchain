<script lang="ts" setup>
import {
  faBolt,
  faCog,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faNetworkWired,
} from '@fortawesome/pro-solid-svg-icons'
import type { HashTabs } from '~/types/hashTabs'

const { t: $t } = useTranslation()

const visible = defineModel<boolean>()

const tabs: HashTabs = [
  {
    icon: faCog,
    key: 'general',
    title: $t('notifications.tabs.general'),
  },
  {
    icon: faGaugeSimpleMax,
    key: 'dashboards',
    title: $t('notifications.tabs.dashboards'),
  },
  {
    icon: faMonitorWaveform,
    key: 'machines',
    title: $t('notifications.tabs.machines'),
  },
  {
    icon: faBolt,
    key: 'clients',
    title: $t('notifications.tabs.clients'),
  },
  {
    icon: faNetworkWired,
    key: 'networks',
    title: $t('notifications.tabs.networks'),
  },
]
</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('notifications.management.title')"
    class="notifications-management-modal-container"
  >
    <div id="notifications-management-search-placholder" />
    <BcTabList
      :tabs
      default-tab="summary"
      class="notifications-management-tab-view"
      oanels-class="notifications-management-tab-panels"
    >
      <template #tab-panel-general>
        <NotificationsManagementGeneralTab />
      </template>
      <template #tab-panel-dashboards>
        <NotificationsManagementDashboards />
      </template>
      <template #tab-panel-machines>
        <NotificationsManagementMachines />
      </template>
      <template #tab-panel-clients>
        <NotificationsManagementClients />
      </template>
      <template #tab-panel-networks>
        <NotificationsManagementNetwork />
      </template>
    </BcTabList>
    <Button
      class="done-button"
      :label="$t('navigation.done')"
      @click="visible = false"
    />
  </BcDialog>
</template>

<style lang="scss" scoped>
#notifications-management-search-placholder {
  position: absolute;
  top: 70px;
  right: var(--padding-large);
  z-index: 2;

  @media (max-width: 1100px) {
    top: var(--padding-large);
  }
}

:global(.notifications-management-modal-container) {
  position: relative;
  width: 1400px;
  height: 786px;
}

:global(.notifications-management-modal-container .p-dialog-header) {
  margin-right: 40px;
}

.notifications-management-tab-view {
  margin-top: var(--padding-large);

  :deep(.notifications-management-panels) {
    min-height: 652px;
  }
}

.done-button {
  position: absolute;
  bottom: calc(var(--padding-large) + var(--padding));
  right: calc(var(--padding-large) + var(--padding));
}
</style>
