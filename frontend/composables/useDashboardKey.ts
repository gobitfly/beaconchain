import { inject } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'

export function useDashboardKey() {
  const { isLoggedIn } = useUserStore()

  const data = inject<DashboardKeyData>('dashboard-key')
  const hasGuestDashboardKeyChanged = ref(false)

  if (!data) {
    throw new Error(
      'useDashboardKey must be in a child of useDashboardKeyProvider',
    )
  }

  const dashboardKey = computed(() => data.dashboardKey.value ?? '')
  const isGuestDashboard = computed(() => !!data.isGuestDashboard.value)
  const isSharedDashboard = computed(() => !!data.isSharedDashboard.value)
  const publicEntities = computed(() => data.publicEntities.value ?? [])
  const isPrivateDashboard = computed(() => isLoggedIn.value && !isGuestDashboard.value)
  const setDashboardKey = (key: string) => data.setDashboardKey(key)
  const dashboardType = computed(() => data.dashboardType.value)

  watch(() => [
    dashboardKey.value,
    isGuestDashboard.value,
  ] as const, (newValues, oldValues) => {
    // Since guest dashboards change their key depending on the validators they contain,
    // sometimes we need to track if the key change is because of this reason.
    const [
      ,isGuestDashboard,
    ] = newValues
    const [
      prevDashboardKey,
      wasGuestDashboard,
    ] = oldValues

    hasGuestDashboardKeyChanged.value = !!prevDashboardKey && wasGuestDashboard && isGuestDashboard
  })

  return {
    ...data,
    dashboardKey,
    dashboardType,
    hasGuestDashboardKeyChanged,
    isGuestDashboard,
    isPrivateDashboard,
    isSharedDashboard,
    publicEntities,
    setDashboardKey,
  }
}
