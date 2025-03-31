import type {
  InternalGetLatestStateResponse,
  LatestStateData,
} from '~/types/api/latest_state'

export const useLatestStateStore = defineStore('latestState', () => {
  const { fetch } = useCustomFetch()

  const latestState = ref<LatestStateData>()

  async function refreshLatestState() {
    try {
      const res = await fetch<InternalGetLatestStateResponse>(
        'LATEST_STATE',
      )
      if (!res.data) {
        return null
      }
      latestState.value = res.data
    }
    catch {
      return null
    }
  }
  return {
    latestState,
    refreshLatestState,
  }
})
