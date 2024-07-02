<script lang="ts" setup>
import {
  faEdit
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type Menubar from 'primevue/menubar'
import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import BcTooltip from '../bc/BcTooltip.vue'
import type { MenuBarButton, MenuBarEntry } from '~/types/menuBar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { type Dashboard, type CookieDashboard, COOKIE_DASHBOARD_ID, type DashboardType, type DashboardKey } from '~/types/dashboard'
import { DashboardRenameModal } from '#components'

const { width } = useWindowSize()
const dialog = useDialog()

interface Props {
  dashboardTitle?: string,
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const route = useRoute()
const router = useRouter()
const isValidatorDashboard = route.name === 'dashboard-id'
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const { isLoggedIn } = useUserStore()
const { refreshOverview } = useValidatorDashboardOverviewStore()
const { dashboards, getDashboardLabel, refreshDashboards } = useUserDashboardStore()
const { dashboardKey, dashboardType, setDashboardKey, isPrivate, isShared } = useDashboardKey()

const emit = defineEmits<{(e: 'showCreation'): void }>()

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

  const sortedItems: MenuBarButton[][] = []

  const addToSortedItems = (minButtonCount: number, items?: MenuBarButton[]) => {
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
  const createMenuBarButton = (type: DashboardType, label: string, id: DashboardKey): MenuBarButton => {
    if (type === dashboardType.value) {
      return { label, command: () => setDashboardKey(id), active: id === dashboardKey.value, route: `/dashboard/${id}` }
    }

    if (type === 'validator') {
      return { label, route: `/dashboard/${id}` }
    }
    return { label, route: `/account-dashboard/${id}` }
  }
  addToSortedItems(0, dashboards.value?.validator_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return createMenuBarButton('validator', getDashboardName(cd), `${cd.hash !== undefined ? cd.hash : cd.id}`)
  }))
  addToSortedItems(3, dashboards.value?.account_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return createMenuBarButton('account', getDashboardName(cd), `${cd.hash ?? cd.id}`)
  }))
  const disabledTooltip = !showInDevelopment ? $t('common.coming_soon') : undefined
  const onNotificaitonsPage = dashboardType.value === 'notifications'
  addToSortedItems(2, [{ label: $t('dashboard.notifications'), route: !onNotificaitonsPage ? '/notifications' : undefined, disabledTooltip, active: onNotificaitonsPage }])

  // if we are in a public dashboard and change the validators then the route does not get updated
  const fixedRoute = router.resolve({ name: route.name!, params: { id: dashboardKey.value } })
  return sortedItems.map((items) => {
    const active = items.find(i => i.active || i.route === fixedRoute.path)
    return {
      active: !!active,
      label: active?.label ?? items[0].label,
      dropdown: items.length > 1,
      disabledTooltip: items.length === 1 ? items[0].disabledTooltip : undefined,
      route: items.length === 1 ? items[0].route : undefined,
      command: items.length === 1 ? items[0].command : undefined,
      items: items.length > 1 ? items : undefined
    }
  })
})

const title = computed(() => {
  return props?.dashboardTitle || getDashboardLabel(dashboardKey.value, isValidatorDashboard ? 'validator' : 'account')
})

const editDashboard = () => {
  const list = isValidatorDashboard ? dashboards.value?.validator_dashboards : dashboards.value?.account_dashboards
  const dashboard = list?.find(d => `${d.id}` === dashboardKey.value)
  if (!dashboard) {
    return
  }
  dialog.open(DashboardRenameModal, {
    data: {
      dashboard,
      dashboardType: dashboardType.value
    },
    onClose: (value?: DynamicDialogCloseOptions | undefined) => {
      if (value?.data === true) {
        refreshDashboards()
        refreshOverview(dashboardKey.value)
      }
    }
  })
}
</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ title }}
    </div>
    <Button v-if="isPrivate" class="p-button-icon-only edit_button" @click="editDashboard">
      <FontAwesomeIcon :icon="faEdit" />
    </Button>
    <div class="dashboard-buttons">
      <Menubar :class="menuBarClass" :model="items" breakpoint="0px">
        <template #item="{ item }">
          <BcTooltip v-if="item.disabledTooltip" :text="item.disabledTooltip" class="button-content" @click.stop.prevent="() => undefined">
            <span class="text-disabled">{{ item.label }}</span>
          </BcTooltip>
          <BcLink v-else-if="item.route && !item.command" :to="item.route" class="pointer" :class="{ 'p-active': item.active }">
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
        <IconPlus alt="Plus icon" width="100%" height="100%" />
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--padding);
  margin-bottom: var(--padding-large);
  gap: var(--padding);

  .dashboard-title {
    @include utils.truncate-text;
  }

  .edit_button {
    border-color: var(--container-border-color);
    background-color: var(--container-background);
    color: var(--container-color);
    flex-shrink: 0;
  }

  .dashboard-buttons {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    flex-grow: 1;
    gap: var(--padding);
    justify-content: flex-end;

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

    :deep(.p-menubar-root-list > .p-menuitem) {
      width: 145px;
    }
    :deep(.p-menubar-root-list .p-menuitem) {

      &:has(>.p-menuitem-content .text-disabled) {
          cursor: default;
        > .p-menuitem-content{
          opacity: 0.5;
        }
      }
    }
  }
}
</style>
