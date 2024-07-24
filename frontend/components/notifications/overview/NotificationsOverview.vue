<script setup lang="ts">
import { defineProps } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDiscord, faTwitter, faGithub } from '@fortawesome/free-brands-svg-icons'
import { faDesktop } from '@fortawesome/pro-solid-svg-icons'
import type { NotificationsOverview } from '~/types/notifications/overview'

const props = defineProps<{ store: NotificationsOverview | null }>()
console.log(props + 'this is props.store')

const {t: $t} = useI18n()
const tPath = 'notifications.overview.'

// const labels = computed(() => {
//   return [
//     { label: $t(`${tPath}email_notifications`), value: props.store?.EmailNotifications ? 'Active' : 'not Active' },
//     { label: $t(`${tPath}push_notifications`), value: props.store?.pushNotifications ? 'Active' : 'not Active' },
//     { label: $t(`${tPath}most_notifications_30d`), value: props.store?.mostNotifications30d ?? 0 },
//     { label: $t(`${tPath}most_notificatoins_24h`), value: props.store?.mostNotifications24h ?? 0 }
//   ]
// })

</script>

<template>
  <div class="container">
    <!-- <div v-for="label in labels" :key="label.label">
    </div> -->
    <div>
      <div v-if="props.store" class="box">
        <div class="box-item">
          <span class="big_text">Email Notifications:</span>
          <span class="big_text_label">
            {{ props.store.EmailNotifications ? 'Active' : 'not Active' }}
          </span>
        </div>
        <div class="box-item">
          <span class="big_text">Push Notifications:</span>
          <span class="big_text_label">
            {{ props.store.pushNotifications ? 'Active' : 'not Active' }}
          </span>
        </div>
        <div class="box-item">
          <span>Most Notifications in 30 Days:</span>
          <span>{{ props.store.mostNotifications30d ? props.store.mostNotifications30d : '0' }}</span>
          <span><ol><li><icon><faDesktop /></icon> Hetzner</li></ol></span>
        </div>
        <div class="box-item">
          <span>Most Notifications in 24 Hours:</span>
          <span>{{ props.store.mostNotifications24h ? props.store.mostNotifications24h : '0' }}</span>
          <span>10 Email | 100 Webhook | 100 Push</span> 
        </div>
      </div>
      <div v-else>
        No data available from the component.
      </div>
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
</style>
