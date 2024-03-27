import { inject } from 'vue'
import type { DateInfo } from '~/types/date'

/**
* @param: tickSeconds: if set then the tickTimestamp will be updated with a new timestamp every x seconds
**/
export function useDate () {
  const date = inject<DateInfo>('date-info')

  if (!date) {
    throw new Error('useDate must be in a child of useDateProvider')
  }

  return date
}
