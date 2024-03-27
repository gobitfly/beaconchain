import { DateTime } from 'luxon'
import { inject } from 'vue'
import type { DateInfo } from '~/types/date'

/**
* @param: tickSeconds: if set then the tickTimestamp will be updated with a new timestamp every x seconds
**/
export function useDate (tickSeconds?: number) {
  const date = inject<DateInfo>('date-info')
  const tickTimestamp = ref(date?.timestamp.value)

  if (!date) {
    throw new Error('useDate must be in a child of useDateProvider')
  }

  watch(date?.timestamp, (ts) => {
    if (!tickSeconds) {
      return
    }
    const dt = DateTime.fromMillis(ts)
    if (!tickTimestamp.value || dt.diff(DateTime.fromMillis(tickTimestamp.value), 'seconds').seconds >= tickSeconds) {
      tickTimestamp.value = ts
    }
  })

  return { ...date, tickTimestamp }
}
