import {
  type WatchDebouncedOptions,
  watchDebounced as watchDebouncedVueUse,
} from '@vueuse/core'
import type { WatchCallback } from 'vue'

export function watchDebounced<
  T extends object,
  Immediate extends Readonly<boolean> = false,
>(source: T, cb: WatchCallback<T, Immediate extends true ? T | undefined : T>,
  options?: WatchDebouncedOptions<Immediate>,
) {
  return watchDebouncedVueUse(
    source,
    cb,
    {
      debounce: 500,
      maxWait: 1000,
      ...options,
    },
  )
}
