export const currentTimestampInSeconds = () => Math.round(Date.now() / 1000)

export const getFutureTimestampInSeconds = (
  {
    seconds = 0,
  } = {}) => Math.round(Date.now() / 1000) + seconds

export const getSeconds = (
  {
    hours = 0,
    minutes = 0,
    seconds = 0,
  }: {
    hours?: number,
    minutes?: number,
    seconds?: number,
  },
) => {
  const hoursInSeconds = hours * 60 * 60
  const minutesInSeconds = minutes * 60
  return hoursInSeconds + minutesInSeconds + seconds
}

export const formatSecondsTo = (seconds: number,
  {
    locale = 'en-US',
    maximumFractionDigits = 2,
    minimumFractionDigits = 2,
    minimumIntegerDigits = 1,
  }:
  {
    locale?: string,
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
