import { defineStore } from 'pinia'
import type { LatestState } from '~/types/latestState'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useLatestStateStore = defineStore('latest-state-store', () => {
  const latest = ref<LatestState | undefined | null>()
  async function getLatestState () {
    if (process.server) {
      const res = await useCustomFetch<LatestState>(API_PATH.LATEST_STATE)
      latest.value = res
    } else {
      // TODO remove this once we can load the data also from the client
    }
    return latest.value
  }

  return { latest, getLatestState }
})
