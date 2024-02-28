import { inject } from 'vue'
import type { WindowSize } from '~/types/window'

export function useWindowSize () {
  const size = inject<WindowSize>('windowSize')

  const width = computed(() => size?.width?.value ?? 2000)
  const height = computed(() => size?.height?.value ?? 2000)

  const isMobile = computed(() => {
    return (width.value ?? 0) <= 800
  })

  const isIPad = computed(() => {
    return (width.value ?? 0) <= 1194
  })

  return { width, height, isMobile, isIPad }
}
