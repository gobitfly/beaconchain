<script setup lang="ts">
import { defineProps } from 'vue'
import type { NotificationsOverview } from '~/types/notifications/overview'

const props = defineProps<{ store: NotificationsOverview | null }>()
console.log(props + 'this is props.store')

const {t: $t} = useI18n()

</script>

<template>
  <div class="container">
    <div class="info-section">
      <span class="icon">🔔</span>
      <span class="text">Notification Overview</span>
    </div>
    <div>
      <div v-if="props.store" class="box">
        <div class="box-item"><span class="big_text">Email Notifications:</span> <span class="big_text_label">{{ props.store.EmailNotifications }}</span></div>
        <div class="box-item"><span>Push Notifications:</span><span>{{ props.store.pushNotifications }}</span> </div>
        <div class="box-item"><span>Most Notifications in 30 Days:</span><span>{{ props.store.mostNotifications30d }}</span> </div>
        <div class="box-item"><span>Most Notifications in 24 Hours:</span><span>{{ props.store.mostNotifications24h }}</span> </div>
      </div>
      <div v-else>
        No data available from the component.
      </div>
    </div>
    <div class="action-section">
      <span class="icon">⚙️</span>
      <span class="text">Settings</span>
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
</style>
