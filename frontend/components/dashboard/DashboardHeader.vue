<script lang="ts" setup>
import type Menubar from 'primevue/menubar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import type { DashboardKey } from '~/types/dashboard'

const { width } = useWindowSize()

interface Props {
  dashboardKey?: DashboardKey // optional because it's not available for notifications
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const store = useUserDashboardStore()
const { getDashboards } = store

const { dashboards } = storeToRefs(store)
await useAsyncData('validator_dashboards', () => getDashboards()) // TODO: This is called here and in DashboardValidatorManageValidators.vue. Should just be called once?

const emit = defineEmits<{(e: 'showCreation'): void }>()

interface MenuBarButton {
  label: string;
  active: boolean;
  route?: string;
  class?: string;
}

interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean;
  items?: MenuBarButton[];
}

const items = computed(() => {
  if (dashboards.value === undefined) {
    return []
  }

  const dashboardsButtons: MenuBarEntry[] = []

  let buttonCount = 3 // [validator], [accounts], [notifications]
  if (width.value < 680) {
    if (width.value < 550) {
      buttonCount = 1 // [validator, accounts, notifications]
    } else {
      buttonCount = 2 // [validator, accounts], [notifications]
    }
  }

  const sortedItems: MenuBarButton[][] = []

  const addToSortedItems = (minButtonCount: number, items?:MenuBarButton[]) => {
    if (items?.length) {
      if (buttonCount >= minButtonCount) {
        sortedItems.push(items)
      } else {
        let last = sortedItems.length - 1
        if (last < 0) {
          sortedItems.push([])
          last = 0
        } else {
          sortedItems[last][sortedItems[last].length - 1].class = 'p-big-separator'
        }
        sortedItems[last] = sortedItems[last].concat(items)
      }
    }
  }
  addToSortedItems(0, dashboards.value?.validator_dashboards.map(({ id, name }) => ({ label: name || `${$t('dashboard.validator_dashboard')} ${id}`, active: id === props.dashboardKey, route: `/dashboard/${id}` })))
  addToSortedItems(3, dashboards.value?.account_dashboards.map(({ id, name }) => ({ label: name || `${$t('dashboard.account_dashboard')} ${id}`, active: id === props.dashboardKey, route: `/dashboard/${id}` })))
  addToSortedItems(2, [{ label: $t('dashboard.notifications'), active: props.dashboardKey === undefined, route: '/notifications' }])

  for (const items of sortedItems) {
    let activeLabel = ''
    items.forEach((item) => {
      if (item.active) {
        activeLabel = item.label
      }
    })

    dashboardsButtons.push({
      label: activeLabel !== '' ? activeLabel : items[0].label,
      active: activeLabel !== '',
      dropdown: items.length > 1,
      route: items.length === 1 ? items[0].route : undefined,
      items: items.length > 1 ? items : undefined
    })
  }

  return dashboardsButtons
})
</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ $t('dashboard.title') }}
    </div>
    <div class="dashboard-buttons">
      <Menubar :model="items" breakpoint="0px">
        <template #item="{ item }">
          <NuxtLink v-if="item.route" :to="item.route">
            <span class="button-content" :class="[item.class, {'p-active': item.active, 'pointer': item.dropdown}]">
              <span class="text">{{ item.label }}</span>
              <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
            </span>
          </NuxtLink>
          <span v-else class="button-content" :class="{ 'p-active': item.active, 'pointer': item.dropdown }">
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
