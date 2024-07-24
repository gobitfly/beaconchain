<script setup lang="ts">
import { defineProps, computed } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDesktop, faUser } from '@fortawesome/pro-solid-svg-icons'
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
  props.store?.mostNotifications30d ?? { providers: [], abo: [] }
)
const mostNotifications24h = computed(() => 
  props.store?.mostNotifications24h ?? { Email: 0, Webhook: 0, Push: 0 }
)

// Computed property to calculate the total notifications
const totalNotifications24h = computed(() => {
  const notifications = mostNotifications24h.value
  return notifications.Email + notifications.Webhook + notifications.Push
})
</script>

<template>
  <div class="container">
    <div v-if="props.store" class="box">
      <div class="box-item">
        <span class="big_text_label">Email Notifications:</span>
        <span class="big_text">{{ emailNotificationStatus }}</span>
        <div v-if="emailNotificationStatus === 'Inactive'" class="small_text">
          Click<a href="/notifications"> here </a> to activate
        </div>
      </div>
      <div class="box-item">
        <span class="big_text_label">Push Notifications:</span>
        <span class="big_text">{{ pushNotificationStatus }}</span>
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
  align-items: center;    /* Center items horizontally */
  justify-content: space-between; /* Center items vertically if they do not exceed the container height */
  gap: 100px;
  align-content: center;
}

.box-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

a:hover {
  color: var(--light-blue);
}

.lists-container {
  display: flex; /* Aligns child elements (the ol elements) in a row */
  gap: 20px; /* Space between the lists */
}

.icon-list {
  list-style-type: none; /* Remove default list numbers */
  padding: 0; /* Remove default padding */
  margin: 0; /* Remove default margin */
  display: flex; /* Align list items horizontally */
  flex-direction: column; /* Stack list items vertically */
  gap: 10px; /* Space between list items */
}

.icon-list li {
  display: flex; /* Align items within the list item horizontally */
  align-items: center; /* Center icon and text vertically */
  gap: 10px; /* Space between the icon and text */
}

.icon {
  font-size: 16px; /* Adjust icon size if necessary */
}
</style>
