import type { StringUnitLength } from 'luxon'
import type { AgeFormat } from '~/types/settings'

export function useFormat() {
  const {
    getTimestampFromEpoch,
    getTimestampFromSlot,
  } = useNetworkStore()

  function formatEpochToDateTime(
    epoch: number,
    timestamp?: number,
    format?: AgeFormat,
    style?: StringUnitLength,
    locales?: string,
    withTime?: boolean,
  ): null | string | undefined {
    return formatTs(
      getTimestampFromEpoch(epoch),
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
    return formatTs(
      getTimestampFromSlot(slot),
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
    return formatEpochToDateTime(
      epoch,
      undefined,
      'absolute',
      undefined,
      locales,
      false,
    )
  }

  return {
    formatEpochToDate,
    formatEpochToDateTime,
    formatSlotToDateTime,
  }
}
