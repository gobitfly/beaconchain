<script lang="ts" setup>
import { useNotificationsManagementSettings } from '~/composables/notifications/useNotificationsManagementSettings'

const { t: $t } = useTranslation()

const visible = defineModel<boolean>()
const { pairedDevices } = useNotificationsManagementSettings()

</script>

<template>
  <BcDialog
    v-model="visible"
    class="paired-devices-modal-container"
  >
    <div class="container">
      <h1>{{ $t('notifications.general.paired_devices.title') }}</h1>
      <div class="paired-devices">
        <NotificationsManagementPairedDeviceModalContent v-for="device in pairedDevices" :key="device.id" :device="device" />
      </div>
    </div>
    <div class="button-row">
      <Button :label="$t('navigation.done')" @click="visible = false" />
    </div>
  </BcDialog>
</template>

<style lang="scss" scoped>
:global(.paired-devices-modal-container) {
  width: 790px;
}

.container {
  padding: var(--padding-large);

  h1 {
    margin-top: 0;
  }

  .paired-devices {
    display: flex;
    flex-direction: column;
    gap: var(--padding-large);

    >:not(:last-child) {
      padding-bottom: var(--padding-large);
      border-bottom: 1px solid var(--container-border-color);
    }
  }
}

.button-row {
  margin-top: var(--padding);
  display: flex;
  justify-content: flex-end;
}
</style>
