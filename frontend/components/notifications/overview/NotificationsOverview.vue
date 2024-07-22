<script setup lang="ts">
import { defineProps } from 'vue'
import type { NotificationsOverview } from '~/types/notifications/overview'

const props = defineProps<{ store: NotificationsOverview | null }>()
console.log(props + 'this is props.store')

const {t: $t} = useI18n()

</script>

<template>
  <div class="container">
    <div>
      <div v-if="props.store" class="box">
        <div><span class="big_text">Email Notifications:</span> <span class="big_text_label">{{ props.store.EmailNotifications }}</span></div>
        <div>Push Notifications: {{ props.store.pushNotifications }}</div>
        <div>Most Notifications in 30 Days: {{ props.store.mostNotifications30d }}</div>
        <div>Most Notifications in 24 Hours: {{ props.store.mostNotifications24h }}</div>
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
  overflow-x: auto;
  gap: 50px;
  height: 101px;
  padding-left: var(--padding-xl);
  padding-right: var(--padding-xl);
}

.box {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
