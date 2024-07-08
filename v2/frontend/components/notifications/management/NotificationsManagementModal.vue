<script lang="ts" setup>
import {
  faCog,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faBolt,
  faNetworkWired
} from '@fortawesome/pro-solid-svg-icons'
import { useUseNotificationsManagementGeneralProvider } from '~/composables/notifications/useNotificationsManagementGeneralProvider'

const { t: $t } = useI18n()

const visible = defineModel<boolean>()

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const { refreshGeneralSettings } = useUseNotificationsManagementGeneralProvider()
await refreshGeneralSettings()

</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('notifications.management.title')"
    class="notifications-management-modal-container"
  >
    <div id="notifications-management-search-placholder" />
    <TabView lazy class="notifications-management-tab-view">
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.general')" :icon="faCog" />
        </template>
        <NotificationsManagementGeneralTab />
      </TabPanel>
      <TabPanel :disabled="!showInDevelopment">
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.dashboards')" :icon="faGaugeSimpleMax" />
        </template>
        <NotificationsManagementDashboards />
      </TabPanel>
      <TabPanel :disabled="!showInDevelopment">
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.machines')" :icon="faMonitorWaveform" />
        </template>
        Machines coming soon!
      </TabPanel>
      <TabPanel :disabled="!showInDevelopment">
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.clients')" :icon="faBolt" />
        </template>
        Clients coming soon!
      </TabPanel>
      <TabPanel :disabled="!showInDevelopment">
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.rocketpool')">
            <template #icon>
              <IconRocketPool />
            </template>
          </BcTabHeader>
        </template>
        Rocket Pool coming soon!
      </TabPanel>
      <TabPanel :disabled="!showInDevelopment">
        <template #header>
          <BcTabHeader :header="$t('notifications.tabs.network')" :icon="faNetworkWired" />
        </template>
        Network coming soon!
      </TabPanel>
    </TabView>
    <Button class="done-button" :label="$t('navigation.done')" @click="visible = false" />
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

:global(.notifications-management-tab-view >.p-tabview-panels) {
  min-height: 652px;
}

.p-tabview {
  margin-top: var(--padding-large);
}

.done-button {
  position: absolute;
  bottom: calc(var(--padding-large) + var(--padding));
  right: calc(var(--padding-large) + var(--padding));
}
</style>
