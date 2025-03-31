import type { StringUnitLength } from 'luxon'
import type { Locale } from '~/i18n/i18n.config'
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
    locales?: Locale,
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
    locales?: Locale,
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
    locales: Locale,
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
