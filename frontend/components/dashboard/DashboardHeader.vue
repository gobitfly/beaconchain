<script lang="ts" setup>
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
const { t: $t } = useI18n()
const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards()) // TODO: This is called here and in DashboardValidatorManageValidators.vue. Should just be called once?

const emit = defineEmits<{(e: 'showCreation'): void }>()

const items = computed(() => {
  interface DashboardButton {
    label: string;
    items?: {
        label: string;
    }[];
  }

  const dashboardsButtons: DashboardButton[] = []

  let items = dashboards.value?.validator_dashboards.map(({ name }) => ({ label: name }))
  dashboardsButtons.push({
    label: 'Validators',
    items
  })

  items = dashboards.value?.account_dashboards.map(({ name }) => ({ label: name }))
  dashboardsButtons.push({
    label: 'Accounts',
    items
  })

  dashboardsButtons.push({
    label: $t('dashboard.notifications')
  })

  return dashboardsButtons
})

</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ $t('dashboard.title') }}
    </div>
    <div class="dashboard-buttons">
      <Menubar :model="items" />
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
