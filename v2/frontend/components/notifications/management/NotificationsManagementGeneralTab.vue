<script lang="ts" setup>
// TODO: Implement Do not disturb feature (mind new design)
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

const { generalSettings, updateGeneralSettings } = useNotificationsManagementSettings()

const pairedDevicesModalVisible = ref(false)
const doNotDisturbToggle = ref(false)
const emailToggle = ref(false)
const pushToggle = ref(false)
const { value: testButtonsDisabled, bounce: bounceTestButton, instant: setTestButton } = useDebounceValue<boolean>(false, 5000)

const sendTestNotification = async (type: 'email' | 'push') => {
  setTestButton(true)
  if (type === 'email') {
    await fetch(API_PATH.NOTIFICATIONS_TEST_EMAIL)
  } else {
    await fetch(API_PATH.NOTIFICATIONS_TEST_PUSH)
  }
  bounceTestButton(false)
}

const pairedDevices = computed(() => generalSettings?.value?.paired_devices?.length || 0)

const openPairdeDevicesModal = () => {
  pairedDevicesModalVisible.value = true
}

watch(generalSettings, (g) => {
  if (g) {
    emailToggle.value = g.enable_email
    pushToggle.value = g.enable_push
  }
}, { immediate: true })

watch([emailToggle, pushToggle], ([enableEmail, enablePush]) => {
  if (!generalSettings.value) {
    return
  }
  if (generalSettings.value?.enable_email !== enableEmail || generalSettings.value?.enable_push !== enablePush) {
    updateGeneralSettings({ ...generalSettings.value, enable_email: enableEmail, enable_push: enablePush })
  }
})

</script>

<template>
  <NotificationsManagementPairedDevicesModal v-model="pairedDevicesModalVisible" />
  <div class="container">
    <div class="row divider">
      <div>
        <span>{{ $t('notifications.general.do_not_disturb') }}</span>
        <span class="explanation">Mutes all notifications</span>
      </div>
      <BcToggle v-model="doNotDisturbToggle" />
    </div>
    <div class="row">
      <div>
        {{ $t('notifications.general.email_notifications') }}
      </div>
      <BcToggle v-model="emailToggle" />
    </div>
    <div class="row divider">
      <div>
        {{ $t('notifications.general.push_notifications') }}
        <span v-if="pairedDevices > 0">
          ({{ pairedDevices }})
          <FontAwesomeIcon
            class="link popout"
            :icon="faArrowUpRightFromSquare"
            @click="openPairdeDevicesModal"
          />
        </span>
      </div>
      <BcToggle v-if="pairedDevices > 0" v-model="pushToggle" />
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

.container{
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

    .explanation{
      @include fonts.tiny_text;
      color: var(--text-color-discreet);
      margin-left: var(--padding-small);
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
</style>
