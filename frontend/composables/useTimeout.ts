export function useTimeout (ms: number) {
  const refreshTimeout = ref<NodeJS.Timeout | null>(null)
  const tick = ref<number>(0)

  onMounted(() => {
    refreshTimeout.value = setInterval(() => { tick.value = new Date().getTime() }, ms)
  })
  onUnmounted(() => {
    refreshTimeout.value && clearInterval(refreshTimeout.value)
  })
  return { tick }
}
