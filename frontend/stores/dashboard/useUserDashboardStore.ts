import { defineStore } from 'pinia'
import { warn } from 'vue'
import type {
  GetUserDashboardsResponse,
  UserDashboardsData,
  ValidatorDashboard,
} from '~/types/api/dashboard'
import type { VDBPostReturnData } from '~/types/api/validator_dashboard'
import {
  COOKIE_DASHBOARD_ID,
  type CookieDashboard,
  type DashboardKey,
  type DashboardType,
} from '~/types/dashboard'
import { COOKIE_KEY } from '~/types/cookie'
import { API_PATH } from '~/types/customFetch'
import type { ChainIDs } from '~/types/network'
import {
  isPublicDashboardKey, isSharedKey,
} from '~/utils/dashboard/key'

const userDashboardStore = defineStore('user_dashboards_store', () => {
  const data = ref<null | undefined | UserDashboardsData>()
  return { data }
})

export function useUserDashboardStore() {
  const { fetch } = useCustomFetch()
  const { t: $t } = useTranslation()
  const { data } = storeToRefs(userDashboardStore())
  const { isLoggedIn } = useUserStore()
  const dashboardCookie = useCookie(COOKIE_KEY.USER_DASHBOARDS)

  const dashboards = computed(() => data.value)

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
      data.value = res.data

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
      data.value = cookieDashboards.value
    }
    return dashboards.value
  }

  // Public dashboards are saved in a cookie (so that it's accessable during SSR)
  function saveToCookie(db: null | undefined | UserDashboardsData) {
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
  ): Promise<CookieDashboard | undefined> {
    if (!isLoggedIn.value) {
      // Create local Validator dashboard
      const cd: CookieDashboard = {
        hash: dashboardKey ?? '',
        id: COOKIE_DASHBOARD_ID.VALIDATOR,
        name: '',
      }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [ cd as ValidatorDashboard ],
      }
      saveToCookie(data.value)
      return cd
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
      data.value = {
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
  ): Promise<CookieDashboard | undefined> {
    if (!isLoggedIn.value) {
      // Create local account dashboard
      const cd: CookieDashboard = {
        hash: dashboardKey ?? '',
        id: COOKIE_DASHBOARD_ID.ACCOUNT,
        name: '',
      }
      data.value = {
        account_dashboards: [ cd ],
        validator_dashboards: dashboards.value?.validator_dashboards || [],
      }
      saveToCookie(data.value)
      return cd
    }
    // Create user specific account dashboard
    const res = await fetch<{ data: VDBPostReturnData }>(
      API_PATH.DASHBOARD_CREATE_ACCOUNT,
      { body: { name } },
    )
    if (res.data) {
      data.value = {
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

  // Update the hash (=hashed list of id's) of a specific local dashboard
  function updateHash(type: DashboardType, hash: string) {
    if (!isPublicDashboardKey(hash) || isSharedKey(hash)) {
      warn('invalid public hashed key: ', hash)
      return
    }
    if (type === 'validator') {
      const cd: CookieDashboard = {
        id: COOKIE_DASHBOARD_ID.VALIDATOR,
        name: '',
        ...dashboards.value?.validator_dashboards?.[0],
        hash,
      }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [ cd as ValidatorDashboard ],
      }
    }
    else {
      const cd: CookieDashboard = {
        id: COOKIE_DASHBOARD_ID.ACCOUNT,
        name: '',
        ...dashboards.value?.account_dashboards?.[0],
        hash,
      }
      data.value = {
        account_dashboards: [ cd ],
        validator_dashboards: dashboards.value?.validator_dashboards || [],
      }
    }
    saveToCookie(data.value)
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
    updateHash,
  }
}
