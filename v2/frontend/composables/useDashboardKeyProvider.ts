import { pullAll, union } from 'lodash-es'
import { provide, warn } from 'vue'
import { COOKIE_KEY } from '~/types/cookie'
import type { DashboardKey, DashboardKeyData, DashboardType } from '~/types/dashboard'
import { isPublicDashboardKey, isSharedKey } from '~/utils/dashboard/key'
export function useDashboardKeyProvider (type: DashboardType = 'validator', mockKey: DashboardKey = '') {
  const route = useRoute()
  const router = useRouter()
  const dashboardType = ref(type)
  const dashboardKey = ref(mockKey)
  const dashboardKeyCookie = useCookie(dashboardType.value === 'account' ? COOKIE_KEY.ACCOUNT_DASHOBARD_KEY : COOKIE_KEY.VALIDATOR_DASHOBARD_KEY)
  const { isLoggedIn } = useUserStore()

  const setDashboardKey = (key: string) => {
    if (!route.name) {
      warn('route name missing', route)
    }
    const newRoute = router.resolve({ name: route.name!, params: { id: key }, hash: document?.location?.hash })
    dashboardKey.value = key
    if (process.client) {
      // we only want to change the url in the browser and don't want to trigger a page refresh
      history.replaceState({}, '', newRoute.fullPath)
    }
    dashboardKeyCookie.value = dashboardKey.value
  }

  const initialCheck = () => {
    if (mockKey) {
      return
    }
    if (!route.params.id && dashboardKey.value !== undefined) {
      if (!dashboardKeyCookie.value) {
        return
      }
      // only use the dashboard cookie key as default if you are not logged in and it's not private
      if (!isLoggedIn.value && isPublicDashboardKey(dashboardKeyCookie.value) && !isSharedKey(dashboardKeyCookie.value)) {
        setDashboardKey(`${dashboardKeyCookie.value}`)
      }
      return
    }
    if (Array.isArray(route.params.id)) {
      setDashboardKey(toBase64Url(route.params.id.join(',')))
    } else {
      setDashboardKey(route.params.id)
    }
  }
  initialCheck()

  const isPublic = computed(() => {
    return isPublicDashboardKey(dashboardKey.value)
  })

  const isShared = computed(() => {
    return isSharedKey(dashboardKey.value)
  })

  // validator id / publicKey for validator dashboard or account id or ens name for account dashboard
  const publicEntities = computed(() => {
    if (!isPublic.value || !dashboardKey.value) {
      return []
    }
    return fromBase64Url(dashboardKey.value)?.split(',') ?? []
  })

  const updateEntities = (list:string[]) => {
    const filtered = list.filter(s => !!s).join(',')
    const key = toBase64Url(filtered)
    setDashboardKey(key)
  }

  const addEntities = (list:string[]) => {
    updateEntities(union(publicEntities.value, list))
  }

  const removeEntities = (list:string[]) => {
    updateEntities(pullAll(publicEntities.value, list))
  }

  const api = { dashboardKey, isPublic, isShared, publicEntities, addEntities, removeEntities, setDashboardKey, dashboardType }

  watch(isLoggedIn, (newValue, oldValue) => {
    if (oldValue && !newValue && dashboardKeyCookie.value && !isNaN(parseInt(dashboardKeyCookie.value))) {
      setDashboardKey('')
    }
  })

  provide<DashboardKeyData>('dashboard-key', api)
  return api
}
