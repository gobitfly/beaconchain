import { pullAll, union } from 'lodash-es'
import { provide, warn } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'
export function useDashboardKeyProvider () {
  const route = useRoute()
  const router = useRouter()
  const dashboardKey = ref('')

  watch(() => route, (r) => {
    console.log('watch route', r, router.resolve({ name: r.name!, params: { id: 'test' } }))
    if (Array.isArray(r.params.id)) {
      dashboardKey.value = toBase64Url(r.params.id.join(','))
    } else {
      dashboardKey.value = r.params.id
    }
  }, { immediate: true })

  const isPublic = computed(() => {
    const id = parseInt(dashboardKey.value)
    return isNaN(id)
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
    const newRoute = router.resolve({ name: route.name!, params: { id: key } })
    dashboardKey.value = key
    console.log('new entities', key, list, newRoute.fullPath)
    // we only want to change the url in the browser and don't want to trigger a page refresh
    history.replaceState({}, '', newRoute.fullPath)
  }

  const addEntities = (list:string[]) => {
    updateEntities(union(publicEntities.value, list))
  }

  const removeEntities = (list:string[]) => {
    updateEntities(pullAll(publicEntities.value, list))
  }
  const api = { dashboardKey, isPublic, publicEntities, addEntities, removeEntities }

  provide<DashboardKeyData>('dashboard-key', api)
  return api
}
