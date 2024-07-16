<script lang="ts" setup>
// TODO: Translations
// TODO: Get linked app "Beaconchain Dashboard" from NotificationSettingsPairedDevice
// TODO: Get paired time stamp from NotificationSettingsPairedDevice
import { faTrash } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type NotificationSettingsPairedDevice } from '~/types/notifications/settings'

interface Props {
  device: NotificationSettingsPairedDevice
}
const props = defineProps<Props>()

const notificationsToggle = ref(false)
const accountAccessToggle = ref(false)

</script>

<template>
  <div class="row-container">
    <div class="device-row ">
      <div class="device">
        Device: {{ props.device.name || 'Unknown' }}
      </div>
      <Button secondary class="p-button-icon-only">
        <FontAwesomeIcon :icon="faTrash" />
      </Button>
    </div>
    <div class="linked-app-row">
      Linked App: Beaconchain Dashboard
    </div>
    <div class="toggle-row">
      <BcToggle v-model="notificationsToggle" />
      Mobile Notifications
    </div>
    <div class="toggle-row">
      <BcToggle v-model="accountAccessToggle" />
      Grant account access (required)
    </div>
    <div class="paired-row">
      Paired: 27 minutes ago
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

  .linked-app-row {
    @include fonts.small_text;
    margin-bottom: var(--padding);
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
