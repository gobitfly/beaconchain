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
  dropdown: boolean;
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
      dropdown: items.length > 1,
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
      dropdown: items.length > 1,
      items
    })
  }

  dashboardsButtons.push({
    label: $t('dashboard.notifications'),
    active: false, // TODO: Active handling missing
    dropdown: false
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
      <Menubar :model="items">
        <template #item="{ item }">
          <span class="button-content" :class="{ 'p-active': item.active, 'pointer': item.dropdown }">
            <span class="text">{{ item.label }}</span>
            <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
          </span>
        </template>
      </Menubar>
      <Button class="p-button-icon-only" @click="emit('showCreation')">
        <IconPlus alt="Plus icon" width="100%" height="100%" />
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

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

    .button-content{
      display: flex;
      align-items: center;
      justify-content: space-between;

      .text {
        @include utils.truncate-text;
      }

      .toggle {
        flex-shrink: 0;
      }

      .pointer {
        cursor: pointer;
      }
    }

    :deep(.p-menubar-root-list > .p-menuitem) {
      width: 130px;
    }
  }
}
</style>
