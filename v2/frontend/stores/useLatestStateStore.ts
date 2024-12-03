import { defineStore } from 'pinia'
import type {
  InternalGetLatestStateResponse,
  LatestStateData,
} from '~/types/api/latest_state'

const latestStateStore = defineStore('latest_state_store', () => {
  const data = ref<LatestStateData | null | undefined>()
  return { data }
})

export function useLatestStateStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(latestStateStore())

  const latestState = computed(() => data.value)

  async function refreshLatestState() {
    try {
      const res = await fetch<InternalGetLatestStateResponse>(
        'LATEST_STATE',
      )
      if (!res.data) {
        return null
      }
      data.value = res.data
      return data.value
    }
    catch {
      return null
    }
  }

  return {
    latestState,
    refreshLatestState,
  }
}
