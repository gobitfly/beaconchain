/**
 * Usage:
 *
 * Case 1. Let us say that we want to debounce a text input so that the API is not called more often than once per second.
 * In <script setup> we can write something like:
 *   `const debouncer = useDebounceValue<string>('', 1000)`
 *   `watch(debouncer.value, callTheAPIwhenItIsTime)`
 * Like so, our function `callTheAPIwhenItIsTime()` will get called with minimum intervals of 1000 ms.
 * To inform the debouncer that it can start counting time, we do:
 *   `debouncer.bounce(input.value, false, true)`
 * each time that our input changes. The last parameter `true` restarts the timer after each key stroke, so your function
 * will not get called while the user types, only when they do a break longer than 1 second.
 *
 * Case 2. Let us say that we want to debounce a button, for which it makes sense to get clicked several times in a row
 * (you simply do not want a flood of requests from someone clicking way too fast).
  * In <script setup> we can write something like:
 *   `const debouncer = useDebounceValue<type identifying the button>('', 300)`
 *   `watch(debouncer.value, doSomethingwhenItIsTime)`
 * Like so, our function `doSomethingwhenItIsTime()` will get called with minimum intervals of 300 ms (a good click rate for a human who wants to skip information).
 * To inform the debouncer that it can start counting time, we do:
 *   `debouncer.bounce(data identifying the button)`
 * each time that our button is clicked.
 */
export function useDebounceValue<T> (initialValue: T, bounceMs: number = 100) {
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

  const bounce = (value: T, instantIfNoTimer = false, endlesBounce = false) => {
    tempRef.value = value
    if (instantIfNoTimer && !timeout.value) {
      valueRef.value = value
    }
    if (endlesBounce || !timeout.value) {
      removeTimeout()
      timeout.value = setTimeout(timeFinished, bounceMs)
    }
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
