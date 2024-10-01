<script lang="ts" setup>
import { faTrash } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type NotificationPairedDevice } from '~/types/api/notifications'

const { t: $t } = useTranslation()

const props = defineProps<{
  device: NotificationPairedDevice,
}>()

const emit = defineEmits<{
  (e: 'remove-device', id: string): void,
  (e: 'toggle-notifications', {
    id,
    value,
  }: {
    id: string,
    value: boolean,
  }): void,
}>()

const hasNotifications = ref(props.device.is_notifications_enabled)
</script>

<template>
  <div class="row-container">
    <div class="device-row">
      <div class="device truncate-text">
        {{ $t("notifications.general.paired_devices.device") }}:
        {{ device.name || $t("notifications.general.paired_devices.unknown") }}
      </div>
      <Button
        severity="secondary"
        class="p-button-icon-only margin-inline-start-small"
        @click="emit('remove-device', device.id)"
      >
        <FontAwesomeIcon :icon="faTrash" />
      </Button>
    </div>
    <div class="toggle-row">
      <BcToggle
        v-model="hasNotifications"
        @update:model-value="emit('toggle-notifications', { id: device.id, value: $event })"
      />
      {{ $t("notifications.general.paired_devices.mobile_notifications") }}
    </div>
    <div class="paired-row">
      {{
        $t("notifications.general.paired_devices.paired_date", {
          date: formatGoTimestamp(device.paired_timestamp),
        })
      }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

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
  .margin-inline-start-small {
    margin-inline-start: var(--padding-small);
  }
}
</style>
