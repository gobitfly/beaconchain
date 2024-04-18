<script lang="ts" setup>
import type Menubar from 'primevue/menubar'
import type { MenuBarButton, MenuBarEntry } from '~/types/menuBar'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import type { Dashboard } from '~/types/api/dashboard'
import { type CookieDashboard, COOKIE_DASHBOARD_ID } from '~/types/dashboard'

const { width } = useWindowSize()

const { t: $t } = useI18n()
const route = useRoute()
const router = useRouter()
const isValidatorDashboard = route.name === 'dashboard-id'

const { isLoggedIn } = useUserStore()
const { dashboards } = useUserDashboardStore()
const { dashboardKey } = useDashboardKey()
const { overview } = useValidatorDashboardOverviewStore()

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

const getDashboardName = (db: Dashboard):string => {
  if (isLoggedIn.value) {
    return db.name || `${$t('dashboard.title')} ${db.id}` // Just to be sure, we should not have dashboards without a name in prod
  } else {
    return db.id === COOKIE_DASHBOARD_ID.ACCOUNT ? $t('dashboard.account_dashboard') : $t('dashboard.validator_dashboard')
  }
}

const items = computed<MenuBarEntry[]>(() => {
  if (dashboards.value === undefined) {
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
  addToSortedItems(0, dashboards.value?.validator_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return { label: getDashboardName(cd), route: `/dashboard/${cd.hash ?? cd.id}` }
  }))
  addToSortedItems(3, dashboards.value?.account_dashboards?.map((db) => {
    const cd = db as CookieDashboard
    return { label: getDashboardName(cd), route: `/account/${cd.hash ?? cd.id}` }
  }))
  addToSortedItems(2, [{ label: $t('dashboard.notifications'), route: '/notifications' }])

  return sortedItems.map((items) => {
    // if we are in a public dashboard and change the validators then the route does not get updated
    const fixedRoute = router.resolve({ name: route.name!, params: { id: dashboardKey.value } })
    const active = items.find(i => i.route === fixedRoute.path)
    return {
      active: !!active,
      label: active?.label ?? items[0].label,
      dropdown: items.length > 1,
      route: items.length === 1 ? items[0].route : active?.route,
      items: items.length > 1 ? items : undefined
    }
  })
})

const title = computed(() => {
  const list = isValidatorDashboard ? dashboards.value?.validator_dashboards : dashboards.value?.account_dashboards
  const id = parseInt(dashboardKey.value ?? '')
  if (!isNaN(id)) {
    const userDb = list?.find(db => db.id === id)
    if (userDb) {
      return userDb.name
    }
    // in production we should not get here, but with our public api key we can also view dashboards that are not part of our list
    if (overview.value) {
      return `${isValidatorDashboard ? $t('dashboard.validator_dashboard') : $t('dashboard.account_dashboard')} ${id}`
    }
  }
  const cookieDb = (list as CookieDashboard[])?.find(db => db.hash === dashboardKey.value)
  if (cookieDb || (isLoggedIn.value && !dashboardKey.value)) {
    return isValidatorDashboard ? $t('dashboard.validator_dashboard') : $t('dashboard.account_dashboard')
  }

  return isValidatorDashboard ? $t('dashboard.public_validator_dashboard') : $t('dashboard.public_account_dashboard')
})

</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ title }}
    </div>
    <div class="dashboard-buttons">
      <Menubar :class="menuBarClass" :model="items" breakpoint="0px">
        <template #item="{ item }">
          <NuxtLink v-if="item.route" :to="item.route" :class="{ 'p-active': item.active }">
            <span class="button-content" :class="[item.class, { 'pointer': item.dropdown }]">
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
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--padding-large);

  .dashboard-title {
    @include utils.truncate-text;
  }

  .dashboard-buttons {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: var(--padding);

    .button-content {
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
