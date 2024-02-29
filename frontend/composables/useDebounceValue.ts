export function useDebounceValue<T> (initialValue: T, bounceMs: number = 100) {
  const valueRef = shallowRef<T>(initialValue)
  const value = readonly(valueRef)
  const tempRef = shallowRef<T>(initialValue)
  const temp = readonly(valueRef)
  const timeout = ref<NodeJS.Timeout | null>(null)

  const removeTimeout = () => {
    timeout.value && clearTimeout(timeout.value)
    timeout.value = null
  }

  const timeFinished = () => {
    valueRef.value = tempRef.value
    timeout.value = null
  }

  const bounce = (value: T, instantIfNoTimer = false) => {
    tempRef.value = value
    if (instantIfNoTimer) {
      if (!timeout.value) {
        valueRef.value = value
        timeout.value = setTimeout(timeFinished, bounceMs)
      }
      return
    }
    removeTimeout()
    timeout.value = setTimeout(timeFinished, bounceMs)
  }

  const instant = (value: T) => {
    tempRef.value = value
    valueRef.value = value
  }

  onUnmounted(() => {
    removeTimeout()
  })
  return { value, temp, bounce, instant }
}
