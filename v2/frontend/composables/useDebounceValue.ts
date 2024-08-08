export function useDebounceValue<T>(initialValue: T, bounceMs: number = 100) {
  const valueRef = shallowRef<T>(initialValue)
  const value = readonly(valueRef)
  const tempRef = shallowRef<T>(initialValue)
  const temp = readonly(tempRef)
  const timeout = ref<NodeJS.Timeout | null>(null)

  const removeTimeout = () => {
    timeout.value && clearTimeout(timeout.value)
    timeout.value = null
  }

  const timeFinished = () => {
    valueRef.value = tempRef.value
    timeout.value = null
  }

  const bounce = (
    value: T,
    instantIfNoTimer = false,
    endlesBounce = false,
    ms?: number,
  ) => {
    tempRef.value = value
    if (instantIfNoTimer && !timeout.value) {
      valueRef.value = value
    }
    if (endlesBounce || !timeout.value) {
      removeTimeout()
      timeout.value = setTimeout(timeFinished, ms || bounceMs)
    }
  }

  const instant = (value: T) => {
    tempRef.value = value
    valueRef.value = value
  }

  onUnmounted(() => {
    removeTimeout()
  })
  return { bounce, instant, temp, value }
}
