import type { Locale } from '~/i18n/i18n.config'

export const currentTimestampInSeconds = () => Math.round(Date.now() / 1000)

export const getFutureTimestampInSeconds = (
  {
    seconds = 0,
  } = {}) => Math.round(Date.now() / 1000) + seconds

export const getSeconds = (
  {
    days = 0,
    hours = 0,
    minutes = 0,
    seconds = 0,
  }: {
    days?: number,
    hours?: number,
    minutes?: number,
    seconds?: number,
  },
) => {
  const daysInSeconds = days * 24 * 60 * 60
  const hoursInSeconds = hours * 60 * 60
  const minutesInSeconds = minutes * 60
  return daysInSeconds + hoursInSeconds + minutesInSeconds + seconds
}

export const formatSecondsTo = (seconds: number,
  {
    locale = 'en-US',
    maximumFractionDigits = 2,
    minimumFractionDigits = 2,
    minimumIntegerDigits = 1,
  }:
  {
    locale?: Locale,
    maximumFractionDigits?: number,
    minimumFractionDigits?: number,
    minimumIntegerDigits?: number,
  } = {},
) => {
  const format
  = (value: number) => {
    return new Intl.NumberFormat(locale, {
      maximumFractionDigits,
      minimumFractionDigits,
      minimumIntegerDigits,
    })
      .format(value)
  }
  const minutes = format(seconds / 60)
  return {
    minutes,
  }
}

export const getRelativeTime = (unixTimestamp: number, {
  locale = 'en-US',
  style = 'short',
}: {
  locale?: Locale,
  style?: 'long' | 'short',
} = {}) => {
  const seconds = unixTimestamp - (Date.now() / 1000)
  const minutes = (seconds / 60)
  const hours = (minutes / 60)
  const days = (hours / 24)
  const weeks = (days / 7)

  const formatter = new Intl.RelativeTimeFormat(locale, {
    style,
  })

  if (Math.abs(weeks) >= 1) {
    return formatter.format(Math.round(weeks), 'weeks')
  }
  if (Math.abs(days) >= 1) {
    return formatter.format(Math.round(days), 'days')
  }
  if (Math.abs(hours) >= 1) {
    return formatter.format(Math.round(hours), 'hours')
  }
  if (Math.abs(minutes) >= 1) {
    return formatter.format(Math.round(minutes), 'minutes')
  }
  return formatter.format(Math.round(seconds), 'seconds')
}

export const getDateTime = (unixTimestamp: number, {
  hasDate = true,
  hasTime = true,
  locale = 'en-US',
}: {
  hasDate?: boolean,
  hasTime?: boolean,
  locale?: Locale,
} = {}) => {
  return new Intl.DateTimeFormat(locale, {
    day: hasDate ? 'numeric' : undefined,
    hour: hasTime ? 'numeric' : undefined,
    minute: hasTime ? 'numeric' : undefined,
    month: hasDate ? 'short' : undefined,
    year: hasDate ? 'numeric' : undefined,
  }).format(unixTimestamp * 1000)
}
