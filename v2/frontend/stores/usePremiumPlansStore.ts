import { defineStore } from 'pinia'
import type { PremiumPlanAPIresponse } from '~/types/pricing'
import { API_PATH } from '~/types/customFetch'

const premiumPlansStore = defineStore('premium_plans_store', () => {
  const data = ref < PremiumPlanAPIresponse>()

  return { data }
})

export function usePremiumPlansStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(premiumPlansStore())

  const premiumPlans = computed(() => data.value)

  async function getPremiumPlans () {
    if (data.value) {
      return data.value
    }

    const res = await fetch<PremiumPlanAPIresponse>(API_PATH.PREMIUM_PLANS)

    data.value = res
    return res
  }

  return { premiumPlans, getPremiumPlans }
}
