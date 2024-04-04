import { inject } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'

export function useDashboardKey () {
  const data = inject<DashboardKeyData>('dashboard-key')

  if (!data) {
    throw new Error('useDashboardKey must be in a child of useDashboardKeyProvider')
  }

  const dashboardKey = computed(() => data.dashboardKey.value ?? '')
  const isPublic = computed(() => !!data.isPublic.value)
  const publicEntities = computed(() => data.publicEntities.value ?? [])

  return { ...data, dashboardKey, isPublic, publicEntities }
}
