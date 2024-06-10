import { type StringUnitLength } from 'luxon'
import * as formatTs from '~/utils/format'
import type { AgeFormat } from '~/types/settings'

export function useFormat () {
  const { currentNetwork } = useNetwork()

  function formatEpochToDateTime (epoch: number, timestamp?: number, format?: AgeFormat, style?: StringUnitLength, locales?: string, withTime?: boolean) : string | null | undefined {
    return formatTs.formatEpochToDateTime(currentNetwork.value, epoch, timestamp, format, style, locales, withTime)
  }

  function formatSlotToDateTime (slot: number, timestamp?: number, format?: AgeFormat, style?: StringUnitLength, locales?: string, withTime?: boolean) : string | null | undefined {
    return formatTs.formatSlotToDateTime(currentNetwork.value, slot, timestamp, format, style, locales, withTime)
  }

  function formatEpochToDate (epoch: number, locales: string): string | null | undefined {
    return formatTs.formatEpochToDate(currentNetwork.value, epoch, locales)
  }

  return { formatEpochToDateTime, formatSlotToDateTime, formatEpochToDate }
}
