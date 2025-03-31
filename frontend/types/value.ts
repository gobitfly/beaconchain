export type ExtendedLabel = {
  fullLabel?: string,
  label: NumberOrString,
}

export type NumberOrString = number | string

export const TimeFrames = [
  'last_24h',
  'last_7d',
  'last_30d',
  'all_time',
] as const
export type CompareResult = 'equal' | 'higher' | 'lower'

export type TimeFrame = (typeof TimeFrames)[number]
