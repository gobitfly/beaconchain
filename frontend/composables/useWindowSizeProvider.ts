import { ref, provide } from 'vue'
import type { WindowSize } from '~/types/window'

export function useWindowSizeProvider() {
  const width = ref(2000)
  const height = ref(2000)

  const validatePageSize = () => {
    width.value = document.body.clientWidth // clientWidth => window width - scrollbar width
    // we need to use the innerHeight as the body.clientHeight is content
    // debendent and we don't want to have horizontal scrollbars over the whole page anyway.
    height.value = window.innerHeight
  }

  onMounted(() => {
    window.addEventListener('resize', validatePageSize)
    validatePageSize()
  })

  onUnmounted(() => {
    window.removeEventListener('resize', validatePageSize)
  })

  provide<WindowSize>('windowSize', { width, height })
}
