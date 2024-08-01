<script setup lang="ts">
import { defineProps, computed } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDesktop, faUser, faInfoCircle } from '@fortawesome/pro-solid-svg-icons'
import type { NotificationsOverview } from '~/types/notifications/overview'

const props = defineProps<{ store: NotificationsOverview | null }>()

const emailNotificationStatus = computed(() => props.store?.EmailNotifications ? 'Active' : 'Inactive')
const pushNotificationStatus = computed(() => props.store?.pushNotifications ? 'Active' : 'Inactive')
const emailLimitCount = computed(() => props.store?.EmailLimitCount ?? 0)
const mostNotifications30d = computed(() => {
  const notificationsActive = props.store?.EmailNotifications || props.store?.pushNotifications
  if (!notificationsActive) {
    return {
      providers: ['-', '-', '-'],
      abo: ['-', '-', '-']
    }
  }

  const providers = props.store?.mostNotifications30d.providers ?? []
  const abo = props.store?.mostNotifications30d.abo ?? []
  return {
    providers: [...providers, ...Array(3 - providers.length).fill('-')].slice(0, 3),
    abo: [...abo, ...Array(3 - abo.length).fill('-')].slice(0, 3)
  }
})
const mostNotifications24h = computed(() => {
  const notificationsActive = props.store?.EmailNotifications || props.store?.pushNotifications
  return notificationsActive ? props.store?.mostNotifications24h ?? { Email: 0, Webhook: 0, Push: 0 } : { Email: 0, Webhook: 0, Push: 0 }
})
const totalNotifications24h = computed(() => {
  const notifications = mostNotifications24h.value
  return notifications.Email + notifications.Webhook + notifications.Push
})

const tooltipEmail = 'Your current limit is ' + emailLimitCount.value + ' emails per day. Your email limit resets in X hours. Upgrade to premium for more.'
</script>

<template>
  <div class="container">
    <div v-if="props.store" class="box">
      <div class="box-item">
        <span class="big_text_label">Email Notifications:</span>
        <span class="big_text">{{ emailNotificationStatus }}</span>
        <div class="inline-items" v-if="emailNotificationStatus === 'Active'">
          <span class="small_text">{{ emailLimitCount }}/10 per day</span>
          <BcTooltip :text="tooltipEmail" position="top" tooltip-class="tooltip">
            <FontAwesomeIcon :icon="faInfoCircle" />
          </BcTooltip>
          <BcPremiumGem class="gem" />
        </div>
        <div v-if="emailNotificationStatus === 'Inactive'" class="premium-invitation small_text">
          Click <a href="/notifications" class="inline-link">here</a> to activate
        </div>
      </div>
      <div class="box-item">
        <span class="big_text_label">Push Notifications:</span>
        <span class="big_text">{{ pushNotificationStatus }}</span>
        <div v-if="pushNotificationStatus === 'Inactive'" class="push-invitation small_text">
          Download the <a href="/notifications" class="inline-link">mobile app</a> to activate
        </div>
      </div>
      <div class="box-item">
        <span class="big_text_label">Most Notifications in 30 Days:</span>
        <span class="lists-container">
          <ol class="icon-list">
            <li v-for="(provider, index) in mostNotifications30d.providers" :key="'provider-' + index" class="small_text">
              <icon><FontAwesomeIcon :icon="faDesktop" /></icon> {{ provider }}
            </li>
          </ol>
          <ol class="icon-list">
            <li v-for="(abo, index) in mostNotifications30d.abo" :key="'abo-' + index" class="small_text">
              <icon><FontAwesomeIcon :icon="faUser" /></icon> {{ abo }}
            </li>
          </ol>
        </span>
      </div>
      <div class="box-item">
        <span class="big_text_label">Most Notifications in 24 Hours:</span>
        <span class="big_text">{{ totalNotifications24h }}</span>
        <span class="small_text">{{ mostNotifications24h.Email }} Email | {{ mostNotifications24h.Webhook }} Webhook | {{ mostNotifications24h.Push }} Push</span>
      </div>
    </div>
    <div v-else>
      No data available from the component.
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.container {
  @include main.container;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
  padding: 10px 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-section, .action-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.icon {
  font-size: 24px;
}

.text {
  font-size: 18px;
  font-weight: 500;
}

.box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 100px;
  align-content: center;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar {
    display: none; /* Hide scrollbar for WebKit browsers */
  }

  scrollbar-width: none; /* Hide scrollbar for Firefox */
}

.box-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.inline-items {
  display: flex;
  align-items: center;
  gap: 10px;
}

a:hover {
  color: var(--light-blue);
}

.lists-container {
  display: flex;
  gap: 20px;
}

.icon-list {
  list-style-type: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.icon-list li {
  display: flex;
  align-items: center;
  gap: 10px;
}

.icon {
  font-size: 16px;
}

.inline-link,
.gem {
  display: inline-block;
}

.premium-invitation {
  display: flex;
  align-items: center;
  gap: 5px; /* Adjust the gap as needed */
}

.push-invitation {
  display: flex;
  align-items: center;
  gap: 5px; /* Adjust the gap as needed */
}

@media (max-width: 600px) {
  .box {
    flex-direction: row;
    gap: 20px;
  }

  .box-item {
    min-width: 250px; /* Adjust based on content width */
  }
}
</style>
