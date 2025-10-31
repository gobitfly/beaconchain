import { inject } from 'vue'
import type { DateInfo } from '~/types/date'

export function useDate() {
  const date = inject<DateInfo>('date-info')

  if (!date) {
    throw new Error('useDate must be in a child of useDateProvider')
  }

  return date
}
