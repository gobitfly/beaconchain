import { warn } from 'vue'
import type {
  GetUserDashboardsResponse,
  UserDashboardsData,
  ValidatorDashboard,
} from '~/types/api/dashboard'
import type { VDBPostReturnData } from '~/types/api/validator_dashboard'
import {
  type DashboardKey,
  type DashboardType,
  GUEST_DASHBOARD_ID,
  type GuestDashboard,
} from '~/types/dashboard'
import { COOKIE_KEY } from '~/types/cookie'
import { API_PATH } from '~/types/customFetch'
import type { ChainIDs } from '~/types/network'
import {
  isGuestDashboardKey, isSharedDashboardKey,
} from '~/utils/dashboard/key'

export const useUserDashboardStore = defineStore('user_dashboards_store', () => {
  const { fetch } = useCustomFetch()
  const { t: $t } = useTranslation()
  const { isLoggedIn } = useUserStore()
  const dashboardCookie = useCookie(COOKIE_KEY.USER_DASHBOARDS)
  const dashboards = ref<UserDashboardsData>()
  const cookieDashboards = computed(() => {
    if (dashboardCookie.value) {
      if (typeof dashboardCookie.value === 'object') {
        // it seems the browser sometimes auto converts the string into an object
        return dashboardCookie.value as any as UserDashboardsData
      }
      else {
        return JSON.parse(dashboardCookie.value)
      }
    }
  })

  async function refreshDashboards() {
    if (isLoggedIn.value) {
      const res = await fetch<GetUserDashboardsResponse>(
        API_PATH.USER_DASHBOARDS,
      )
      dashboards.value = res.data

      // add fallback names for dashboards that have no names
      if (dashboards.value) {
        dashboards.value.account_dashboards?.forEach((d) => {
          if (d.name === '') {
            d.name = `${$t('dashboard.account_dashboard')} ${d.id}`
          }
        })
        dashboards.value.validator_dashboards?.forEach((d) => {
          if (d.name === '') {
            d.name = `${$t('dashboard.validator_dashboard')} ${d.id}`
          }
        })
      }
    }
    else {
      dashboards.value = cookieDashboards.value
    }
    return dashboards.value
  }

  // Guest dashboards are saved in a cookie (so that it's accessable during SSR)
  function saveToCookie(db: UserDashboardsData) {
    if (isLoggedIn.value) {
      warn('saveToCookie should only be called when not logged in')
      return
    }

    dashboardCookie.value = JSON.stringify(db)
  }

  async function createValidatorDashboard(
    name: string,
    network: ChainIDs,
    dashboardKey?: string,
  ): Promise<GuestDashboard | undefined> {
    if (!isLoggedIn.value) {
      // Create local Validator dashboard
      const gd: GuestDashboard = {
        id: GUEST_DASHBOARD_ID.VALIDATOR,
        key: dashboardKey ?? '',
        name: '',
      }
      dashboards.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [ gd as ValidatorDashboard ],
      }
      saveToCookie(dashboards.value)
      return gd
    }
    // Create user specific Validator dashboard
    const res = await fetch<{ data: VDBPostReturnData }>(
      API_PATH.DASHBOARD_CREATE_VALIDATOR,
      {
        body: {
          name,
          network,
        },
      },
    )
    if (res.data) {
      dashboards.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [
          ...(dashboards.value?.validator_dashboards || []),
          {
            group_count: 1,
            id: res.data.id,
            is_archived: false,
            name: res.data.name,
            network: res.data.network,
            validator_count: 0,
          },
        ],
      }
      return res.data
    }
  }

  async function createAccountDashboard(
    name: string,
    dashboardKey?: string,
  ): Promise<GuestDashboard | undefined> {
    if (!isLoggedIn.value) {
      // Create local account dashboard
      const gd: GuestDashboard = {
        id: GUEST_DASHBOARD_ID.ACCOUNT,
        key: dashboardKey ?? '',
        name: '',
      }
      dashboards.value = {
        account_dashboards: [ gd ],
        validator_dashboards: dashboards.value?.validator_dashboards || [],
      }
      saveToCookie(dashboards.value)
      return gd
    }
    // Create user specific account dashboard
    const res = await fetch<{ data: VDBPostReturnData }>(
      API_PATH.DASHBOARD_CREATE_ACCOUNT,
      { body: { name } },
    )
    if (res.data) {
      dashboards.value = {
        account_dashboards: [
          ...(dashboards.value?.account_dashboards || []),
          {
            id: res.data.id,
            name: res.data.name,
          },
        ],
        validator_dashboards: dashboards.value?.validator_dashboards || [],
      }
      return res.data
    }
  }

  // Update the guest dashboard key (=encoded list of validator indices or public keys) of a specific local dashboard
  function updateGuestDashboardKey(type: DashboardType, key: string) {
    if (!isGuestDashboardKey(key) || isSharedDashboardKey(key)) {
      warn('invalid guest dashboard key: ', key)
      return
    }
    if (type === 'validator') {
      const gd: GuestDashboard = {
        id: GUEST_DASHBOARD_ID.VALIDATOR,
        name: '',
        ...dashboards.value?.validator_dashboards?.[0],
        key,
      }
      dashboards.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [ gd as ValidatorDashboard ],
      }
    }
    else {
      const gd: GuestDashboard = {
        id: GUEST_DASHBOARD_ID.ACCOUNT,
        name: '',
        ...dashboards.value?.account_dashboards?.[0],
        key,
      }
      dashboards.value = {
        account_dashboards: [ gd ],
        validator_dashboards: dashboards.value?.validator_dashboards || [],
      }
    }
    saveToCookie(dashboards.value)
  }

  function getDashboardLabel(key: DashboardKey, type: DashboardType): string {
    const isValidatorDashboard = type === 'validator'
    const list = isValidatorDashboard
      ? dashboards.value?.validator_dashboards
      : dashboards.value?.account_dashboards
    const id = parseInt(key ?? '')
    if (!isNaN(id)) {
      const userDb = list?.find(db => db.id === id)
      if (userDb) {
        return userDb.name
      }

      // in production we should not get here, but with our public api key we
      // can also view dashboards that are not part of our list
      return `${isValidatorDashboard ? $t('dashboard.validator_dashboard') : $t('dashboard.account_dashboard')} ${id}`
    }

    return isValidatorDashboard
      ? $t('dashboard.public_validator_dashboard')
      : $t('dashboard.public_account_dashboard')
  }

  return {
    cookieDashboards,
    createAccountDashboard,
    createValidatorDashboard,
    dashboards,
    getDashboardLabel,
    refreshDashboards,
    updateGuestDashboardKey,
  }
})
