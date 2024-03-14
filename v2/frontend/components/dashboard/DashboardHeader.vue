<script lang="ts" setup>
import type Menubar from 'primevue/menubar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
const { t: $t } = useI18n()
const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards()) // TODO: This is called here and in DashboardValidatorManageValidators.vue. Should just be called once?

const emit = defineEmits<{(e: 'showCreation'): void }>()

interface MenuBarButton {
  label: string;
  active: boolean;
}

interface MenuBarEntry extends MenuBarButton {
  items?: MenuBarButton[];
}

const items = computed(() => {
  // TODO: Test code, should get dashboard key
  const currentDashboardId = '2'

  const dashboardsButtons: MenuBarEntry[] = []

  // TODO: Duplicated code for validators and accounts button
  // Mobile requires special handling, once this is implemented, check whether duplicated code can be reduced
  let items: MenuBarButton[] = dashboards.value?.validator_dashboards.map(({ id, name }) => ({ label: name, active: id === currentDashboardId })) ?? []
  let activeLabel = ''
  items?.forEach((item) => {
    if (item.active) {
      activeLabel = item.label
    }
  })
  if ((items?.length ?? 0) > 0) {
    dashboardsButtons.push({
      label: activeLabel !== '' ? activeLabel : items[0].label,
      active: activeLabel !== '',
      items
    })
  }

  items = dashboards.value?.account_dashboards.map(({ id, name }) => ({ label: name, active: id === currentDashboardId })) ?? []
  activeLabel = ''
  items?.forEach((item) => {
    if (item.active) {
      activeLabel = item.label
    }
  })
  if ((items?.length ?? 0) > 0) {
    dashboardsButtons.push({
      label: activeLabel !== '' ? activeLabel : items[0].label,
      active: activeLabel !== '',
      items
    })
  }

  dashboardsButtons.push({
    label: $t('dashboard.notifications'),
    active: false // TODO: Active handling missing
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
      <div class="dashboard-navigation">
        <Menubar :model="items">
          <template #item="{ item }">
            <span :class="item.active ? 'p-active' : ''">
              {{ item.label }}
            <!--TODO: Dropdown icon-->
            </span>
          </template>
        </Menubar>
      </div>
      <Button class="p-button-icon-only" @click="emit('showCreation')">
        <IconPlus alt="Plus icon" width="100%" height="100%" />
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header-container {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  .dashboard-title {
    margin-bottom: var(--padding-large);
  }

  .dashboard-buttons {
    display: flex;
    align-items: center;
    gap: var(--padding);

    .dashboard-navigation {
      width: calc((3 * 130px) + (2 * var(--padding)));
    }
  }
}
</style>
