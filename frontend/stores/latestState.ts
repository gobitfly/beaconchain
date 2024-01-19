// store/filters.ts
import { defineStore } from 'pinia'
import type { LatestState } from '../types/latestState'

export const useLatestState = defineStore('latest-state-store', () => {
  const { public: { apiClientV1 } } = useRuntimeConfig()

  const latestState = ref<LatestState | undefined>()
  async function getLatestState () {
    if (process.server) {
      const res = await $fetch('/latestState', {
        baseURL: apiClientV1
      })
      latestState.value = res as LatestState
    } else {
      // TODO remove this once we can load the data also from the client
      console.log('we are on the client and have cors issues')
    }
    return latestState.value
  }

  return { latest: latestState, getLatestState }
})
