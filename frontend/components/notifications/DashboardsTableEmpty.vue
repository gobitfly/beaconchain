<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCirclePlus, faRightFromBracket } from '@fortawesome/pro-regular-svg-icons'

const { t: $t } = useTranslation()

const emit = defineEmits<{(e: 'openDialog'): void }>()

const handleClick = () => {
  if (!isLoggedIn.value) {
    return navigateTo('/login')
  }
  if (!hasDashboards.value) {
    return navigateTo('/dashboard')
  }
  emit('openDialog')
}
const { isLoggedIn } = useUserStore()
const { dashboards } = useUserDashboardStore()
const hasDashboards = computed(() => {
  return dashboards.value?.account_dashboards?.length || dashboards.value?.validator_dashboards?.length
})

const text = computed(() => {
  if (!isLoggedIn.value) {
    return $t('notifications.dashboards.empty.login')
  }
  if (!hasDashboards.value) {
    return $t('notifications.dashboards.empty.no_dashboards')
  }
  return $t('notifications.dashboards.empty.with_dashboards')
})

</script>
<template>
  <div class="empty delayed-fadein-animation" @click="handleClick">
    <span class="big_text">{{ text }}</span>
    <FontAwesomeIcon v-if="isLoggedIn" :icon="faCirclePlus" />
    <FontAwesomeIcon v-else :icon="faRightFromBracket" />
  </div>
</template>

<style lang="scss" scoped>
.empty {
  width: 100%;
  height: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: var(--text-color-disabled);
  gap: var(--padding);
  cursor: pointer;
  text-align: center;

  svg {
    width: 30px;
    height: 30px;
  }
}
</style>
