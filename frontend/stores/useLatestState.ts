// store/filters.ts
import { reactive } from 'vue'
import { defineStore } from 'pinia'
import type { LatestState } from '~/types/latestState'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useLatestState = defineStore('latest-state-store', () => {
  const latestState = ref<LatestState | undefined | null>()
  const reactiveState = reactive(latestState)
  async function getLatestState () {
    if (process.server) {
      console.log('store before update', latestState.value)
      const res = await useCustomFetch<LatestState>('/latestState')
      latestState.value = res
      console.log('we are on the client', latestState.value)
    } else {
      // TODO remove this once we can load the data also from the client
      console.log('we are on the client and have cors issues', latestState.value)
    }
    return latestState.value
  }

  return { latest: reactiveState, getLatestState }
})
