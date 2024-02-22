import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { VDBOverviewData, VDBOverviewResponse } from '~/types/api/validator_dashboard'

export const useValidatorDashboardOverview = defineStore('validator_overview', () => {
  const overview = ref<VDBOverviewData | undefined | null>()
  async function getOverview () {
    const res = await useCustomFetch<VDBOverviewResponse>(API_PATH.DASHBOARD_OVERVIEW)
    overview.value = res.data
    return overview.value
  }

  return { overview, getOverview }
})
