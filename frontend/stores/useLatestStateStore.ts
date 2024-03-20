import { defineStore } from 'pinia'
import type { LatestState } from '~/types/latestState'

export const useLatestStateStore = defineStore('latest-state-store', () => {
  const { fetch } = useCustomFetch()
  const latest = ref<LatestState | undefined | null>()
  async function getLatestState () {
    if (process.server) {
      const res = await fetch<LatestState>(API_PATH.LATEST_STATE)
      latest.value = res
    } else {
      // TODO remove this once we can load the data also from the client
    }
    return latest.value
  }

  return { latest, getLatestState }
})
