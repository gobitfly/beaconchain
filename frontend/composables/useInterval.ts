import { DateTime } from 'luxon'

export function useInterval(seconds: number) {
  const { timestamp } = useDate()
  const tick = ref<number>(timestamp.value)
  const internalTick = ref<number>(timestamp.value)

  watch(timestamp, (ts) => {
    if (!seconds) {
      return
    }
    const dt = DateTime.fromMillis(ts)
    // we use in internal tick so that if the interval was reset we don't trigger a change in the tick
    if (
      !internalTick.value
      || dt.diff(DateTime.fromMillis(internalTick.value), 'seconds').seconds
      >= seconds
    ) {
      tick.value = ts
      internalTick.value = ts
    }
  })

  const resetTick = () => {
    internalTick.value = timestamp.value
  }

  return { resetTick, tick }
}
