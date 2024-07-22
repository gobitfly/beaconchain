<script lang="ts" setup>
import { faTrash } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type NotificationPairedDevice } from '~/types/api/notifications'

// TODO: Implement handling of user input

const { t: $t } = useI18n()

interface Props {
  device: NotificationPairedDevice
}
const props = defineProps<Props>()

const notificationsToggle = ref(false)
const accountAccessToggle = ref(false)

</script>

<template>
  <div class="row-container">
    <div class="device-row ">
      <div class="device">
        {{ $t('notifications.general.paired_devices.device') }}: {{ props.device.name || $t('notifications.general.paired_devices.unknown') }}
      </div>
      <Button secondary class="p-button-icon-only">
        <FontAwesomeIcon :icon="faTrash" />
      </Button>
    </div>
    <div class="toggle-row">
      <BcToggle v-model="notificationsToggle" />
      {{ $t('notifications.general.paired_devices.mobile_notifications') }}
    </div>
    <div class="toggle-row">
      <BcToggle v-model="accountAccessToggle" />
      {{ $t('notifications.general.paired_devices.grant_account_access') }}
    </div>
    <div class="paired-row">
      {{ $t('notifications.general.paired_devices.paired_date', {date: formatGoTimestamp(device.paired_timestamp)}) }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.row-container {
  display: flex;
  flex-direction: column;
  gap: var(--padding);

  .device-row {
    @include fonts.subtitle_text;

    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .toggle-row {
    display: flex;
    align-items: center;
    gap: var(--padding-large);
  }

  .paired-row {
    @include fonts.small_text;
    color: var(--text-color-discreet);
    margin-top: var(--padding-small);
  }
}
</style>
