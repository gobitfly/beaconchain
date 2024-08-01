<script lang="ts" setup>
import type Menubar from 'primevue/menubar'
import BcTooltip from '../bc/BcTooltip.vue'
import type { MenuBarButton, MenuBarEntry } from '~/types/menuBar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { type Dashboard, type CookieDashboard, COOKIE_DASHBOARD_ID, type DashboardType, type DashboardKey } from '~/types/dashboard'

const { t: $t } = useI18n()
const { width } = useWindowSize()
const route = useRoute()
const router = useRouter()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const { isLoggedIn } = useUserStore()
const { dashboards } = useUserDashboardStore()
const { dashboardKey, dashboardType, setDashboardKey, isShared } = useDashboardKey()

const emit = defineEmits<{(e: 'showCreation'): void }>()

const getDashboardName = (db: Dashboard): string => {
  if (isLoggedIn.value) {
    return db.name || `${$t('dashboard.title')} ${db.id}` // Just to be sure, we should not have dashboards without a name in prod
  } else {
    return db.id === COOKIE_DASHBOARD_ID.ACCOUNT ? $t('dashboard.account_dashboard') : $t('dashboard.validator_dashboard')
  }
}

const items = computed<MenuBarEntry[]>(() => {
  if (dashboards.value === undefined || isShared.value) {
    return []
  }

  const buttons: MenuBarEntry[] = []

  // if we are in a public dashboard and change the validators then the route does not get updated
  const fixedRoute = router.resolve({ name: route.name!, params: { id: dashboardKey.value } })

  const addToSortedItems = (label: string, items?: MenuBarButton[]) => {
    if (items?.length) {
      const active = items.find(i => i.active || i.route === fixedRoute.path)
      const hasMoreItems = items.length > 1
      const count = hasMoreItems && width.value >= 520 ? ` (${items.length})` : ''
      buttons.push({
        active: !!active,
        label: label + count,
        dropdown: hasMoreItems,
        disabledTooltip: !hasMoreItems ? items[0].disabledTooltip : undefined,
        route: !hasMoreItems ? items[0].route : undefined,
        command: !hasMoreItems ? items[0].command : undefined,
        items: hasMoreItems ? items : undefined
      })
    }
  }
  const createMenuBarButton = (type: DashboardType, label: string, id: DashboardKey): MenuBarButton => {
    if (type === dashboardType.value) {
      return { label, command: () => setDashboardKey(id), active: id === dashboardKey.value, route: `/dashboard/${id}` }
    }

    if (type === 'validator') {
      return { label, route: `/dashboard/${id}` }
    }
    return { label, route: `/account-dashboard/${id}` }
  }
  addToSortedItems($t('dashboard.entity.validator'), dashboards.value?.validator_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return createMenuBarButton('validator', getDashboardName(cd), `${cd.hash !== undefined ? cd.hash : cd.id}`)
  }))
  addToSortedItems($t('dashboard.entity.validator'), dashboards.value?.account_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return createMenuBarButton('account', getDashboardName(cd), `${cd.hash ?? cd.id}`)
  }))
  const disabledTooltip = !showInDevelopment ? $t('common.coming_soon') : undefined
  const onNotificationsPage = dashboardType.value === 'notifications'
  addToSortedItems($t('notifications.title'), [{ label: $t('notifications.title'), route: !onNotificationsPage ? '/notifications' : undefined, disabledTooltip, active: onNotificationsPage }])

  return buttons
})
</script>

<template>
  <div class="header-container">
    <Menubar class="menu-bar" :model="items" breakpoint="0px">
      <template #item="{ item }">
        <BcTooltip
          v-if="item.disabledTooltip"
          :text="item.disabledTooltip"
          class="button-content"
          @click.stop.prevent="() => undefined"
        >
          <span class="text-disabled">{{ item.label }}</span>
        </BcTooltip>
        <BcLink
          v-else-if="item.route && !item.command"
          :to="item.route"
          class="pointer"
          :class="{ 'p-active': item.active }"
        >
          <span class="button-content" :class="[item.class]">
            <span class="text">{{ item.label }}</span>
            <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
          </span>
        </BcLink>
        <span v-else class="button-content pointer" :class="{ 'p-active': item.active }">
          <span class="text">{{ item.label }}</span>
          <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
        </span>
      </template>
    </Menubar>
    <Button v-if="!isShared" class="p-button-icon-only" @click="emit('showCreation')">
      <IconPlus title="Add new dashboard" width="100%" height="100%" />
    </Button>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.header-container {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: var(--padding);
  margin-bottom: var(--padding-large);
  min-width: 1px;
  gap: var(--padding);

  .edit_button {
    border-color: var(--container-border-color);
    background-color: var(--container-background);
    color: var(--container-color);
    flex-shrink: 0;
  }

  .menu-bar {
    display: flex;
    flex-shrink: 1;
    overflow: hidden;

    .button-content {
      display: flex;

      &:has(.toggle) {
        justify-content: space-between;
      }

      .text {
        @include utils.truncate-text;
      }

      .toggle {
        flex-shrink: 0;
        margin-top: auto;
        margin-bottom: auto;
      }

      .pointer {
        cursor: pointer;
      }
    }
  }

  :deep(.p-menubar-root-list > .p-menuitem) {
    width: 145px;
  }

  :deep(.p-menubar-root-list .p-menuitem) {

    &:has(>.p-menuitem-content .text-disabled) {
      cursor: default;

      >.p-menuitem-content {
        opacity: 0.5;
      }
    }
  }

  :deep(.p-menubar-root-list .p-menuitem .p-submenu-list) {
    position: fixed;
  }

  @media (max-width: 519px) {
    gap: var(--padding-small);

    :deep(.p-menubar-root-list) {
      gap: var(--padding-small);
    }

    :deep(.p-menubar-root-list > .p-menuitem > .p-menuitem-content) {
      padding: var(--padding-small) var(--padding);
    }
  }
}
</style>
