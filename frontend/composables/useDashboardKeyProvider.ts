import { pullAll, union } from 'lodash-es'
import { provide, warn } from 'vue'
import type { DashboardKey, DashboardKeyData, DashboardType } from '~/types/dashboard'
export function useDashboardKeyProvider (type: DashboardType = 'validator', mockKey: DashboardKey = '') {
  const route = useRoute()
  const router = useRouter()
  const dashboardKey = ref(mockKey)
  const dashboardKeyCookie = useCookie(`${type}-dashboard-key`)
  const { isLoggedIn } = useUserStore()

  const setDashboardKey = (key: string) => {
    const newRoute = router.resolve({ name: route.name!, params: { id: key } })
    dashboardKey.value = key
    if (process.client) {
      // we only want to change the url in the browser and don't want to trigger a page refresh
      history.replaceState({}, '', newRoute.fullPath)
    } else {
      // if we get here on the server then we have no history
      router.push(newRoute)
    }
  }

  watch(() => route, (r) => {
    if (!r.params.id) {
      if (!dashboardKeyCookie.value) {
        return
      }
      // if you are not logged in then only set the key if it's not an id
      if (isLoggedIn.value || isNaN(parseInt(dashboardKeyCookie.value))) {
        setDashboardKey(dashboardKeyCookie.value)
      }
      return
    }
    if (Array.isArray(r.params.id)) {
      dashboardKey.value = toBase64Url(r.params.id.join(','))
    } else {
      dashboardKey.value = r.params.id
    }
    dashboardKeyCookie.value = dashboardKey.value
  }, { immediate: true })

  const isPublic = computed(() => {
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
    if (!route.name) {
      warn('route name missing', route)
    }
    const key = toBase64Url(list.filter(s => !!s).join(','))
    setDashboardKey(key)
  }

  const addEntities = (list:string[]) => {
    updateEntities(union(publicEntities.value, list))
  }

  const removeEntities = (list:string[]) => {
    updateEntities(pullAll(publicEntities.value, list))
  }

  const api = { dashboardKey, isPublic, publicEntities, addEntities, removeEntities, setDashboardKey }

  provide<DashboardKeyData>('dashboard-key', api)
  return api
}
