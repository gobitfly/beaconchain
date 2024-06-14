import { defineStore } from 'pinia'
import type { InternalGetLatestStateResponse, LatestStateData } from '~/types/api/latest_state'
import { API_PATH } from '~/types/customFetch'

const latestStateStore = defineStore('latest_state_store', () => {
  const data = ref<LatestStateData | undefined | null>()
  return { data }
})

export function useLatestStateStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(latestStateStore())

  const latestState = computed(() => data.value)

  async function refreshLatestState () : Promise<LatestStateData|undefined> {
    try {
      const res = await fetch<InternalGetLatestStateResponse>(API_PATH.LATEST_STATE)
      if (!res.data) {
        return undefined
      }
      data.value = res.data
      return data.value
    } catch {
      return undefined
    }
  }

  return { latestState, refreshLatestState }
}
