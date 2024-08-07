import { type StringUnitLength } from 'luxon'
import type { AgeFormat } from '~/types/settings'
import {
  formatEpochToDateTime as formatEpochToDateTimeImported,
  formatSlotToDateTime as formatSlotToDateTimeImported,
  formatEpochToDate as formatEpochToDateImported,
} from '~/utils/format'

export function useFormat() {
  const { currentNetwork } = useNetworkStore()

  function formatEpochToDateTime(epoch: number, timestamp?: number, format?: AgeFormat, style?: StringUnitLength, locales?: string, withTime?: boolean): string | null | undefined {
    return formatEpochToDateTimeImported(currentNetwork.value, epoch, timestamp, format, style, locales, withTime)
  }

  function formatSlotToDateTime(slot: number, timestamp?: number, format?: AgeFormat, style?: StringUnitLength, locales?: string, withTime?: boolean): string | null | undefined {
    return formatSlotToDateTimeImported(currentNetwork.value, slot, timestamp, format, style, locales, withTime)
  }

  function formatEpochToDate(epoch: number, locales: string): string | null | undefined {
    return formatEpochToDateImported(currentNetwork.value, epoch, locales)
  }

  return { formatEpochToDateTime, formatSlotToDateTime, formatEpochToDate }
}
