<script lang="ts" setup>
const { t: $t } = useTranslation()

const visible = defineModel<boolean>()
const notificationsManagementStore = useNotificationsManagementStore()
const close = () => {
  visible.value = false
}

const handleToggleNotifications = ({
  id,
  value,
}: {
  id: string,
  value: boolean,
}) => {
  notificationsManagementStore.setNotificationForPairedDevice({
    id,
    value,
  })
  const device = notificationsManagementStore.settings.paired_devices.find(device => device.id === id)
  if (device) {
    device.is_notifications_enabled = value
  }
}
</script>

<template>
  <BcDialog
    v-model="visible"
    class="paired-devices-modal-container"
    @keydown.esc.stop.prevent="close"
  >
    <div class="container">
      <h1>{{ $t("notifications.general.paired_devices.title") }}</h1>
      <div v-if="notificationsManagementStore.settings.paired_devices.length" class="paired-devices">
        <NotificationsManagementPairedDeviceModalContent
          v-for="device in notificationsManagementStore.settings.paired_devices"
          :key="device.id"
          :device
          @toggle-notifications="handleToggleNotifications"
          @remove-device="notificationsManagementStore.removeDevice"
        />
      </div>
      <BcText
        v-if="!notificationsManagementStore.settings.paired_devices.length"
        class="info-empty"
        tag="p" variant="lg"
      >
        {{ $t('notifications.general.paired_devices.info_empty.template') }}
        <br>
        <BcLink
          class="link"
          to="/mobile"
        >
          {{ $t('notifications.general.paired_devices.info_empty._link') }}
        </BcLink>
      </BcText>
    </div>
    <div class="button-row">
      <Button
        :label="$t('navigation.done')"
        autofocus
        @click="close"
      />
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

    > :not(:last-child) {
      padding-bottom: var(--padding-large);
      border-bottom: 1px solid var(--container-border-color);
    }
  }
}

.info-empty {
  text-align: center;
}

.button-row {
  margin-top: var(--padding);
  display: flex;
  justify-content: flex-end;
}
</style>
