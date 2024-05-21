import { defineStore } from 'pinia'
import type { InternalGetLatestStateResponse, LatestStateData } from '~/types/api/latest_state'
import { API_PATH } from '~/types/customFetch'

const latestStateStore = defineStore('latest-state-store', () => {
  const data = ref<LatestStateData | undefined | null>()
  return { data }
})

export function useLatestStateStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(latestStateStore())

  const latestState = computed(() => data.value)

  async function refreshLatestState () {
    const res = await fetch<InternalGetLatestStateResponse>(API_PATH.LATEST_STATE)
    data.value = res.data
  }

  return { latestState, refreshLatestState }
}
