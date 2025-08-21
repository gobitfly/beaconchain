<script setup lang="ts">
const { t: $t } = useTranslation()

const emit = defineEmits<{ (e: 'openDialog'): void }>()

const notificationsOverview = useNotificationsOverview()
const {
  validatorDashboards,
} = usePrivateDashboards()

const hasDashboards = computed(() =>
  validatorDashboards.value.length,
)
const hasSubscriptions = computed(() =>
  notificationsOverview.value?.vdb_subscriptions_count
  || notificationsOverview.value?.adb_subscriptions_count,
)
</script>

<template>
  <BcTableEmpty
    :role="hasDashboards ? 'link' : 'button'"
    @click="hasDashboards ? navigateTo('/dashboard') : emit('openDialog')"
  >
    <span v-if="!hasDashboards">{{ $t('notifications.dashboards.empty.no_dashboards') }}</span>
    <span v-else-if="!hasSubscriptions">{{ $t('notifications.dashboards.empty.no_subscriptions') }}</span>
    <span v-else>{{ $t('notifications.dashboards.empty.no_notifications') }}</span>

    <template #icon>
      <BcIcon
        name="circle-plus"
      />
    </template>
  </BcTableEmpty>
</template>
