<script lang="ts" setup>
import type Menubar from 'primevue/menubar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'

const { width } = useWindowSize()

const { t: $t } = useI18n()
const { path } = useRoute()

const { dashboards } = useUserDashboardStore()

const emit = defineEmits<{(e: 'showCreation'): void }>()

interface MenuBarButton {
  label: string;
  route?: string;
  class?: string;
}

interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean;
  items?: MenuBarButton[];
}

const buttonCount = ref<number>(0)
const menuBarClass = ref<string>('')

watch(width, () => {
  menuBarClass.value = ''
  if (width.value < 540) {
    buttonCount.value = 1 // [validator, accounts, notifications]
    menuBarClass.value = 'right-aligned-submenu'
  } else if (width.value < 680) {
    buttonCount.value = 2 // [validator, accounts], [notifications]
  } else {
    buttonCount.value = 3 // [validator], [accounts], [notifications]
  }
}, { immediate: true })

const items = computed<MenuBarEntry[]>(() => {
  if (dashboards.value === undefined) {
    return []
  }

  const sortedItems: MenuBarButton[][] = []

  const addToSortedItems = (minButtonCount: number, items?:MenuBarButton[]) => {
    if (items?.length) {
      if (buttonCount.value >= minButtonCount) {
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
  addToSortedItems(0, dashboards.value?.validator_dashboards?.map(({ id, name }) => ({ label: name || `${$t('dashboard.validator_dashboard')} ${id}`, route: `/dashboard/${id}` })))
  addToSortedItems(3, dashboards.value?.account_dashboards?.map(({ id, name }) => ({ label: name || `${$t('dashboard.account_dashboard')} ${id}`, route: `/account-dashboard/${id}` })))
  addToSortedItems(2, [{ label: $t('dashboard.notifications'), route: '/notifications' }])

  return sortedItems.map((items) => {
    const active = items.find(i => i.route === path)
    return {
      label: active?.label ?? items[0].label,
      dropdown: items.length > 1,
      route: items.length === 1 ? items[0].route : active?.route,
      items: items.length > 1 ? items : undefined
    }
  })
})
</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ $t('dashboard.title') }}
    </div>
    <div class="dashboard-buttons">
      <Menubar :class="menuBarClass" :model="items" breakpoint="0px">
        <template #item="{ item }">
          <NuxtLink v-if="item.route" :to="item.route">
            <span class="button-content" :class="[item.class, { 'pointer': item.dropdown}]">
              <span class="text">{{ item.label }}</span>
              <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
            </span>
          </NuxtLink>
          <span v-else class="button-content" :class="{ 'pointer': item.dropdown }">
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
