<script lang="ts" setup>
import {
  faCog,
  faGaugeSimpleMax,
  faMonitorWaveform,
  faBolt,
  faNetworkWired
} from '@fortawesome/pro-solid-svg-icons'

const { t: $t } = useI18n()

const visible = defineModel<boolean>()
</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.notifications.management.title')"
    class="notifications-management-modal-container"
  >
    <TabView lazy class="notifications-management-tab-view">
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.general')" :icon="faCog" />
        </template>
        General coming soon!
      </TabPanel>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.dashboards')" :icon="faGaugeSimpleMax" />
        </template>
        Dashboards coming soon!
      </TabPanel>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.machines')" :icon="faMonitorWaveform" />
        </template>
        Machines coming soon!
      </TabPanel>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.clients')" :icon="faBolt" />
        </template>
        Clients coming soon!
      </TabPanel>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.rocketpool')">
            <template #icon>
              <IconRocketPool />
            </template>
          </BcTabHeader>
        </template>
        Rocket Pool coming soon!
      </TabPanel>
      <TabPanel>
        <template #header>
          <BcTabHeader :header="$t('dashboard.notifications.management.tabs.network')" :icon="faNetworkWired" />
        </template>
        Network coming soon!
      </TabPanel>
    </TabView>
    <Button class="done-button" :label="$t('navigation.done')" @click="visible=false" />
  </BcDialog>
</template>

<style lang="scss" scoped>
:global(.notifications-management-modal-container) {
  width: 1400px;
  height: 786px;
}

:global(.notifications-management-tab-view >.p-tabview-panels) {
  min-height: 652px;
}

.p-tabview {
  margin-top: var(--padding-large);
}

.done-button{
  position: absolute;
  bottom: calc(var(--padding-large) + var(--padding));
  right: calc(var(--padding-large) + var(--padding));
}

</style>
