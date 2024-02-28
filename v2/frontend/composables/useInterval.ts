export function useInterval (ms: number) {
  const refreshInterval = ref<NodeJS.Timeout | null>(null)
  const tick = ref<number>(0)

  onMounted(() => {
    refreshInterval.value = setInterval(() => { tick.value = new Date().getTime() }, ms)
  })
  onUnmounted(() => {
    refreshInterval.value && clearInterval(refreshInterval.value)
  })
  return { tick }
}
