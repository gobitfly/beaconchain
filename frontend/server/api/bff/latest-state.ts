import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'

export default defineEventHandler(async (event) => {
  return get<InternalGetLatestStateResponse>(event, '/latest-state')
})
