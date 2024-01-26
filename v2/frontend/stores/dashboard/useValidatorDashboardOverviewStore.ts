import { defineStore } from 'pinia'
import type { DashboardOverview } from '~/types/dashboard/overview'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useValidatorDashboardOverview = defineStore('validator_overview', () => {
  const overview = ref<DashboardOverview | undefined | null>()
  async function getOverview () {
    const res = await useCustomFetch<DashboardOverview>(API_PATH.DASHBOARD_OVERVIEW)
    overview.value = res
    return overview.value
  }

  return { overview, getOverview }
})
