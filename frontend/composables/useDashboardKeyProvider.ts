import { pullAll, union } from 'lodash-es'
import { provide } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'
export function useDashboardKeyProvider (area: 'validator' | 'account') {
  const route = useRoute()

  const dashboardKey = computed(() => {
    if (Array.isArray(route.params.id)) {
      return route.params.id.join(',')
    }
    return route.params.id
  })

  const isPublic = computed(() => {
    const id = parseInt(dashboardKey.value)
    return isNaN(id)
  })

  // validator id / publicKey for validator dashboard or account id or ens name for account dashboard
  const publicEntities = computed(() => {
    if (!isPublic.value) {
      return []
    }
    return fromBase64Url(dashboardKey.value)?.split(',') ?? []
  })

  const updateEntities = (value:string[]) => {
    route.params.id = toBase64Url(value.join(','))
  }

  const addEntities = (list:string[]) => {
    updateEntities(union(publicEntities.value, list))
  }

  const removeEntities = (list:string[]) => {
    updateEntities(pullAll(publicEntities.value, list))
  }

  provide<DashboardKeyData>(`${area}-dashboard-key`, { dashboardKey, isPublic, publicEntities, addEntities, removeEntities })
}
