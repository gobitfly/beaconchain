<script setup lang="ts">
import { defineProps, computed } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDesktop } from '@fortawesome/pro-solid-svg-icons'
import type { NotificationsOverview } from '~/types/notifications/overview'

const props = defineProps<{ store: NotificationsOverview | null }>()

const { t: $t } = useI18n()
const tPath = 'notifications.overview.'

// Computed properties for labels and values
const emailNotificationStatus = computed(() => 
  props.store?.EmailNotifications ? 'Active' : 'Inactive'
)
const pushNotificationStatus = computed(() => 
  props.store?.pushNotifications ? 'Active' : 'Inactive'
)
const mostNotifications30d = computed(() => 
  props.store?.mostNotifications30d ?? '0'
)
const mostNotifications24h = computed(() => 
  props.store?.mostNotifications24h ?? '0'
)
</script>

<template>
  <div class="container">
    <div v-if="props.store" class="box">
      <div class="box-item">
        <span class="big_text">Email Notifications:</span>
        <span class="big_text_label">{{ emailNotificationStatus }}</span>
        <div v-if="emailNotificationStatus === 'Inactive'">
          Click<a href="/activate-email-notifications"> here </a> to activate
        </div>
      </div>
      <div class="box-item">
        <span class="big_text">Push Notifications:</span>
        <span class="big_text_label">{{ pushNotificationStatus }}</span>
      </div>
      <div class="box-item">
        <span>Most Notifications in 30 Days:</span>
        <span>{{ mostNotifications30d }}</span>
        <span><ol><li><icon><FontAwesomeIcon :icon="faDesktop" /></icon> Hetzner</li></ol></span>
      </div>
      <div class="box-item">
        <span>Most Notifications in 24 Hours:</span>
        <span>{{ mostNotifications24h }}</span>
        <span>10 Email | 100 Webhook | 100 Push</span>
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
  justify-content: space-between;
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
  gap: 10px;
}

.box-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.big_text {
  font-size: 16px;
  font-weight: bold;
}

.big_text_label {
  font-size: 16px;
}

a:hover {
  color: var(--light-blue);
}
</style>
