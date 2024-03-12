<script lang="ts" setup>
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards()) // TODO: This is called here and in DashboardValidatorManageValidators.vue. Should just be called once?

const emit = defineEmits<{(e: 'showCreation'): void }>()

</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ $t('dashboard.title') }}
    </div>
    <div class="dashboard-buttons">
      <DashboardSelectionButton type="validator" :dashboards="dashboards?.validator_dashboards" />
      <DashboardSelectionButton type="account" :dashboards="dashboards?.account_dashboards" />
      <DashboardSelectionButton type="notifications" />
      <Button class="p-button-icon-only" @click="emit('showCreation')">
        <IconPlus alt="Plus icon" width="100%" height="100%" />
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header-container {
  display: flex;
  justify-content: space-between;

  .dashboard-title {
    margin-bottom: var(--padding-large);
  }

  .dashboard-buttons {
    display: flex;
    gap: var(--padding);
  }
}
</style>
