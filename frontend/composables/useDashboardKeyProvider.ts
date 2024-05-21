import { pullAll, union } from 'lodash-es'
import { provide, warn } from 'vue'
import { COOKIE_KEY } from '~/types/cookie'
import type { DashboardKey, DashboardKeyData, DashboardType } from '~/types/dashboard'
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
    const newRoute = router.resolve({ name: route.name!, params: { id: key } })
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
      // if you are not logged in then only set the key if it's not an id
      if (isLoggedIn.value || isNaN(parseInt(dashboardKeyCookie.value))) {
        setDashboardKey(dashboardKeyCookie.value)
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
    if (!dashboardKey.value) {
      return true
    }

    const id = parseInt(dashboardKey.value)
    return !!dashboardKey.value && isNaN(id)
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

  const api = { dashboardKey, isPublic, publicEntities, addEntities, removeEntities, setDashboardKey, dashboardType }

  watch(isLoggedIn, (newValue, oldValue) => {
    if (oldValue && !newValue && dashboardKeyCookie.value && !isNaN(parseInt(dashboardKeyCookie.value))) {
      setDashboardKey('')
    }
  })

  provide<DashboardKeyData>('dashboard-key', api)
  return api
}
