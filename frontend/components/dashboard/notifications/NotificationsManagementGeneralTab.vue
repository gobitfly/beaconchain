<script lang="ts" setup>
// TODO: Use translations everywhere
// TODO: Re-style toggles (deactivated does not look deactivated right now)
// TODO: Gather all data in on object and call API (use debounce)
// TODO: Hide Push Notifications slider if user das not has the app linked
// TODO: Implement Do not disturb feature
// TOOD: Implement "Paird devices" modal

import {
  faPaperPlane
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const doNotDisturbToggle = ref(false)
const emailToggle = ref(false)
const pushToggle = ref(false)

const sendTestNotification = (type: 'email' | 'push') => {
  alert('TODO: Sending test ' + type + ' notification')
}
</script>

<template>
  <div class="container">
    <div class="row divider">
      <div>
        <span>Do not disturb</span>
        <span class="explanation">Mutes all notifications</span>
      </div>
      <BcToggle v-model="doNotDisturbToggle" />
    </div>
    <div class="row">
      <div>
        E-Mail Notifications
      </div>
      <BcToggle v-model="emailToggle" />
    </div>
    <div class="row divider">
      <div>
        Push Notifications
      </div>
      <BcToggle v-model="pushToggle" />
    </div>
    <div class="row">
      <div>
        Send Test E-Mail
      </div>
      <Button class="p-button-icon-only" @click="sendTestNotification('email')">
        <FontAwesomeIcon :icon="faPaperPlane" />
      </Button>
    </div>
    <div class="row">
      <div>
        Send Test Push Notification
      </div>
      <Button class="p-button-icon-only" @click="sendTestNotification('push')">
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

    &.divider {
      padding-bottom: calc(var(--padding-large) - var(--padding-small));
      margin-bottom: var(--padding-small);
      border-bottom: 1px solid var(--container-border-color);
    }
  }
}
</style>
