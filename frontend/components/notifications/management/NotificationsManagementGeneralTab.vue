<script lang="ts" setup>

import {
  faArrowUpRightFromSquare,
  faPaperPlane
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { API_PATH } from '~/types/customFetch'
import { useNotificationsManagementSettings } from '~/composables/notifications/useNotificationsManagementSettings'
import { Target } from '~/types/links'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

const { generalSettings, updateGeneralSettings, pairedDevices } = useNotificationsManagementSettings()

const isVisible = ref(false)
const isEmailToggleOn = ref(false)
const isPushToggleOn = ref(false)
const { value: testButtonsDisabled, bounce: bounceTestButton, instant: setTestButton } = useDebounceValue<boolean>(false, 5000)

const timestampMute = ref<number | undefined>()
const muteDropdownList = [
  { value: 1 * 60 * 60, label: $t('notifications.general.mute.count_hours', { count: 1 }) },
  { value: 2 * 60 * 60, label: $t('notifications.general.mute.count_hours', { count: 2 }) },
  { value: 4 * 60 * 60, label: $t('notifications.general.mute.count_hours', { count: 4 }) },
  { value: 8 * 60 * 60, label: $t('notifications.general.mute.count_hours', { count: 8 }) },
  { value: Number.MAX_SAFE_INTEGER, label: $t('notifications.general.mute.until_turned_on') }]

const unmuteNotifications = () => {
  timestampMute.value = 0
}

const muteNotifications = (value: number) => {
  if (value === Number.MAX_SAFE_INTEGER) {
    timestampMute.value = Number.MAX_SAFE_INTEGER
    return
  }
  timestampMute.value = (Date.now() / 1000) + value
}

const sendTestNotification = async (type: 'email' | 'push') => {
  setTestButton(true)
  if (type === 'email') {
    await fetch(API_PATH.NOTIFICATIONS_TEST_EMAIL)
  } else {
    await fetch(API_PATH.NOTIFICATIONS_TEST_PUSH)
  }
  bounceTestButton(false)
}

const pairedDevicesCount = computed(() => pairedDevices.value?.length || 0)

const openPairdeDevicesModal = () => {
  isVisible.value = true
}

watch(generalSettings, (newGeneralSettings) => {
  if (newGeneralSettings) {
    isEmailToggleOn.value = newGeneralSettings.is_email_notifications_enabled
    isPushToggleOn.value = newGeneralSettings.is_push_notifications_enabled
    timestampMute.value = newGeneralSettings.do_not_disturb_timestamp > (Date.now() / 1000) ? newGeneralSettings.do_not_disturb_timestamp : undefined
  }
}, { immediate: true })

watch([isEmailToggleOn, isPushToggleOn, timestampMute], ([enableEmail, enablePush, muteTs]) => {
  if (!generalSettings.value) {
    return
  }
  if (generalSettings.value?.is_email_notifications_enabled !== enableEmail || generalSettings.value?.is_push_notifications_enabled !== enablePush || generalSettings.value?.do_not_disturb_timestamp !== muteTs) {
    updateGeneralSettings({ ...generalSettings.value, is_email_notifications_enabled: enableEmail, is_push_notifications_enabled: enablePush, do_not_disturb_timestamp: muteTs! })
  }
})

const textMutedUntil = computed(() => {
  if (timestampMute.value) {
    if (timestampMute.value === Number.MAX_SAFE_INTEGER) {
      return $t('notifications.general.mute.muted_until_turned_on')
    }
    return $t('notifications.general.mute.muted_until_date', { date: formatTsToAbsolute(timestampMute.value, $t('locales.date'), true) })
  }
})

</script>

<template>
  <NotificationsManagementPairedDevicesModal v-model="isVisible" />
  <div class="container">
    <div class="row divider">
      <div>
        <span>{{ $t('notifications.general.do_not_disturb') }}</span>
        <span class="explanation">{{ $t('notifications.general.mute.mutes_all') }}</span>
      </div>
      <div v-if="generalSettings?.do_not_disturb_timestamp" class="unmute-container">
        <Button :label="$t('notifications.general.mute.unmute')" @click="unmuteNotifications()" />
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
        @update:model-value="(value: number)=>muteNotifications(value)"
      >
        <template #value>
          {{ $t('notifications.general.mute.select_duration') }}
        </template>
        <template #option="slotProps">
          {{ slotProps.label }}
        </template>
      </BcDropdown>
    </div>
    <div class="row">
      <div>
        {{ $t('notifications.general.email_notifications') }}
      </div>
      <BcToggle v-model="isEmailToggleOn" />
    </div>
    <div class="row divider">
      <div>
        {{ $t('notifications.general.push_notifications') }}
        <span v-if="pairedDevicesCount > 0">
          ({{ pairedDevicesCount }})
          <FontAwesomeIcon
            class="link popout"
            :icon="faArrowUpRightFromSquare"
            @click="openPairdeDevicesModal"
          />
        </span>
      </div>
      <BcToggle v-if="pairedDevicesCount > 0" v-model="isPushToggleOn" />
      <div v-else>
        {{ tOf($t, 'notifications.general.download_app', 0) }}
        <BcLink to="/mobile  " :target="Target.External" class="link">
          {{ tOf($t, 'notifications.general.download_app', 1) }}
        </BcLink>
        {{ tOf($t, 'notifications.general.download_app', 2) }}
      </div>
    </div>
    <div class="row">
      <div>
        {{ $t('notifications.general.send_test_email') }}
      </div>
      <Button class="p-button-icon-only" :disabled="testButtonsDisabled" @click="sendTestNotification('email')">
        <FontAwesomeIcon :icon="faPaperPlane" />
      </Button>
    </div>
    <div class="row">
      <div>
        {{ $t('notifications.general.send_test_push') }}
      </div>
      <Button class="p-button-icon-only" :disabled="testButtonsDisabled" @click="sendTestNotification('push')">
        <FontAwesomeIcon :icon="faPaperPlane" />
      </Button>
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
