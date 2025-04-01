import { useInterval as useIntervalVueUse } from '@vueuse/core'

export const useInterval = (seconds: number) => {
  const milliSeconds = seconds * 1000
  return useIntervalVueUse(milliSeconds, {
    controls: true,
    immediate: true,
  })
}
