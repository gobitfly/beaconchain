import { DateTime } from 'luxon'

export function useInterval (seconds: number) {
  const { timestamp } = useDate()
  const tick = ref<number>(timestamp.value)

  watch(timestamp, (ts) => {
    if (!seconds) {
      return
    }
    const dt = DateTime.fromMillis(ts)
    if (!tick.value || dt.diff(DateTime.fromMillis(tick.value), 'seconds').seconds >= seconds) {
      tick.value = ts
    }
  })

  return { tick }
}
