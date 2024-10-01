<script lang="ts" setup>
import {
  faArrowUpRightFromSquare,
  faPaperPlane,
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { API_PATH } from '~/types/customFetch'
import { Target } from '~/types/links'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()
const toast = useBcToast()

const notificationsManagementStore = useNotificationsManagementStore()

const isVisible = ref(false)

const muteDropdownList = [
  {
    label: $t('notifications.general.mute.hours', { count: 1 }),
    value: getSeconds({ hours: 1 }),
  },
  {
    label: $t('notifications.general.mute.hours', { count: 2 }),
    value: getSeconds({ hours: 2 }),

  },
  {
    label: $t('notifications.general.mute.hours', { count: 4 }),
    value: getSeconds({ hours: 4 }),

  },
  {
    label: $t('notifications.general.mute.hours', { count: 8 }),
    value: getSeconds({ hours: 8 }),

  },
  {
    label: $t('notifications.general.mute.until_i_turn_on'),
    value: Number.MAX_SAFE_INTEGER,
  },
]

const muteNotifications = (seconds: number) => {
  if (seconds === Number.MAX_SAFE_INTEGER) {
    return notificationsManagementStore
      .settings
      .general_settings
      .do_not_disturb_timestamp = seconds
  }
  notificationsManagementStore
    .settings
    .general_settings
    .do_not_disturb_timestamp = getFutureTimestampInSeconds({ seconds })
}

const sendTestNotification = async (type: 'email' | 'push') => {
  try {
    await fetch(
      type === 'email'
        ? API_PATH.NOTIFICATIONS_TEST_EMAIL
        : API_PATH.NOTIFICATIONS_TEST_PUSH,
    )
  }
  catch (error) {
    toast.showError({
      detail: $t('notifications.general.test_notification_error.toast_message'),
      group: $t('notifications.general.test_notification_error.toast_group'),
      summary: $t('notifications.general.test_notification_error.toast_title'),
    })
  }
}

const pairedDevicesCount = computed(() => notificationsManagementStore.settings.paired_devices?.length || 0)

const hasPushNotificationTest = computed(() =>
  notificationsManagementStore
    .settings
    .general_settings
    .is_push_notifications_enabled
    && notificationsManagementStore.settings.paired_devices?.length,
)

const hasEmailNotificationTest = computed(() =>
  notificationsManagementStore.settings.general_settings.is_email_notifications_enabled,
)
const openPairdeDevicesModal = () => {
  isVisible.value = true
}

const textMutedUntil = computed(() => {
  if (notificationsManagementStore.settings.general_settings.do_not_disturb_timestamp === Number.MAX_SAFE_INTEGER) {
    return $t('notifications.general.mute.until_turned_on')
  }
  return $t('notifications.general.mute.until', {
    date: formatTsToAbsolute(
      notificationsManagementStore.settings.general_settings.do_not_disturb_timestamp,
      $t('locales.date'),
      true,
    ),
  })
})
await notificationsManagementStore.getSettings()
watchDebounced(notificationsManagementStore.settings.general_settings, async () => {
  await notificationsManagementStore.saveSettings()
}, {
  deep: true,
})
</script>

<template>
  <LazyNotificationsManagementPairedDevicesModal
    v-if="isVisible"
    v-model="isVisible"
  />
  <div class="container">
    <div class="row divider do-not-disturb">
      <div>
        <span>{{ $t("notifications.general.do_not_disturb") }}</span>
        <span class="explanation">{{
          $t("notifications.general.mute.all")
        }}</span>
      </div>
      <div
        v-if="notificationsManagementStore.settings.general_settings?.do_not_disturb_timestamp"
        class="unmute-container"
      >
        <Button
          :label="$t('notifications.general.mute.unmute')"
          @click="notificationsManagementStore.settings.general_settings.do_not_disturb_timestamp = 0"
        />
        <div class="muted-until">
          {{ textMutedUntil }}
        </div>
      </div>
      <BcDropdown
        v-else
        :options="muteDropdownList"
        option-value="value"
        option-label="label"
        panel-class="mute-notifications-dropdown-panel"
        @update:model-value="muteNotifications"
      >
        <template #value>
          {{ $t("notifications.general.mute.select_duration") }}
        </template>
        <template #option="slotProps">
          {{ slotProps.label }}
        </template>
      </BcDropdown>
    </div>
    <div class="row">
      <div>
        {{ $t("notifications.general.email_notifications") }}
      </div>
      <BcToggle v-model="notificationsManagementStore.settings.general_settings.is_email_notifications_enabled" />
    </div>
    <div
      class="row"
      :class="{ divider: hasEmailNotificationTest || hasPushNotificationTest }"
    >
      <div>
        {{ $t("notifications.general.push_notifications") }}
        <span v-if="pairedDevicesCount > 0">
          ({{ pairedDevicesCount }})
          <FontAwesomeIcon
            class="link popout"
            :icon="faArrowUpRightFromSquare"
            @click="openPairdeDevicesModal"
          />
        </span>
      </div>
      <BcToggle
        v-if="pairedDevicesCount > 0"
        v-model="notificationsManagementStore.settings.general_settings.is_push_notifications_enabled"
      />
      <div v-else>
        {{ tOf($t, "notifications.general.download_app", 0) }}
        <BcLink
          to="/mobile"
          :target="Target.External"
          class="link"
        >
          {{ tOf($t, "notifications.general.download_app", 1) }}
        </BcLink>
        {{ tOf($t, "notifications.general.download_app", 2) }}
      </div>
    </div>
    <div
      v-if="notificationsManagementStore.settings.general_settings.is_email_notifications_enabled"
      class="row"
    >
      <span>
        {{ $t("notifications.general.send_test_email") }}
      </span>
      <BcButton
        @click="sendTestNotification('email')"
      >
        {{ $t("common.send") }}
        <template #icon>
          <FontAwesomeIcon :icon="faPaperPlane" />
        </template>
      </BcButton>
    </div>
    <div
      v-if="hasPushNotificationTest"
      class="row"
    >
      <span>
        {{ $t("notifications.general.send_test_push") }}
      </span>
      <BcButton
        @click="sendTestNotification('push')"
      >
        {{ $t("common.send") }}
        <template #icon>
          <FontAwesomeIcon :icon="faPaperPlane" />
        </template>
      </BcButton>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.container {
  border: unset;
  margin-top: var(--padding-xl);
  display: flex;
  flex-direction: column;
  align-items: center;
  @include fonts.small_text_bold;

  .row {
    width: 100%;
    max-width: 500px;
    padding: var(--padding);
    display: flex;
    justify-content: space-between;
    align-items: center;

    .explanation {
      @include fonts.tiny_text;
      color: var(--text-color-discreet);
      margin-left: var(--padding-small);
    }

    .unmute-container {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      gap: var(--padding-small);

      .muted-until {
        @include fonts.tiny_text;
        color: var(--text-color-discreet);
      }
    }

    .popout {
      margin-left: var(--padding-small);
    }

    &.divider {
      padding-bottom: calc(var(--padding-large) - var(--padding-small));
      margin-bottom: var(--padding-small);
      border-bottom: 1px solid var(--container-border-color);
    }

    &.do-not-disturb {
      min-height: 76px;
    }
  }
}

:deep(span.p-dropdown-label.p-inputtext) {
  @include fonts.small_text_bold;
}

:global(.mute-notifications-dropdown-panel) {
  li.p-dropdown-item {
    @include fonts.small_text;
  }
}
</style>
