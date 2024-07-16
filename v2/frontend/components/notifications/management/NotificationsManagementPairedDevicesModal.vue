<script lang="ts" setup>
import { useNotificationsManagementSettings } from '~/composables/notifications/useNotificationsManagementSettings'

const { t: $t } = useI18n()

const visible = defineModel<boolean>()
const { generalSettings } = useNotificationsManagementSettings()

</script>

<template>
  <BcDialog
    v-model="visible"
    class="paired-devices-modal-container"
  >
    <div class="container">
      <h1>{{ $t('notifications.general.paired_devices.title') }}</h1>
      <NotificationsManagementPairedDevice v-for="device in generalSettings?.paired_devices" :key="device.id" :device="device" />
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
}

.button-row {
  margin-top: var(--padding);
  display: flex;
  justify-content: flex-end;
}
</style>
