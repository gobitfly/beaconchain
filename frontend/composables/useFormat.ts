import type { StringUnitLength } from 'luxon'
import type { AgeFormat } from '~/types/settings'
import {
  formatEpochToDate as formatEpochToDateImported,
  formatEpochToDateTime as formatEpochToDateTimeImported,
  formatSlotToDateTime as formatSlotToDateTimeImported,
} from '~/utils/format'

export function useFormat() {
  const { currentNetwork } = useNetworkStore()

  function formatEpochToDateTime(
    epoch: number,
    timestamp?: number,
    format?: AgeFormat,
    style?: StringUnitLength,
    locales?: string,
    withTime?: boolean,
  ): null | string | undefined {
    return formatEpochToDateTimeImported(
      currentNetwork.value,
      epoch,
      timestamp,
      format,
      style,
      locales,
      withTime,
    )
  }

  function formatSlotToDateTime(
    slot: number,
    timestamp?: number,
    format?: AgeFormat,
    style?: StringUnitLength,
    locales?: string,
    withTime?: boolean,
  ): null | string | undefined {
    return formatSlotToDateTimeImported(
      currentNetwork.value,
      slot,
      timestamp,
      format,
      style,
      locales,
      withTime,
    )
  }

  function formatEpochToDate(
    epoch: number,
    locales: string,
  ): null | string | undefined {
    return formatEpochToDateImported(currentNetwork.value, epoch, locales)
  }

  return {
    formatEpochToDate,
    formatEpochToDateTime,
    formatSlotToDateTime,
  }
}
