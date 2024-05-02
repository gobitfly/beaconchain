import { defineStore } from 'pinia'
import type { LatestState } from '~/types/latestState'
import { API_PATH } from '~/types/customFetch'

const latestStateStore = defineStore('latest-state-store', () => {
  const data = ref<LatestState | undefined | null>()
  return { data }
})

export function useLatestStateStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(latestStateStore())

  const latestState = computed(() => data.value)

  async function refreshLatestState () {
    if (process.server) {
      const res = await fetch<LatestState>(API_PATH.LATEST_STATE)
      data.value = res
    } else {
      // TODO remove this once we can load the data also from the client
    }
    return latestState.value
  }

  return { latestState, refreshLatestState }
}
