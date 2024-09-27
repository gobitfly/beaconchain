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
