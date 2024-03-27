import { defineStore } from 'pinia'

export const tooltipStore = defineStore('user-store', () => {
  const data = ref<HTMLElement | null>(null)
  return { data }
})

export function useTooltipStore () {
  const { data } = storeToRefs(tooltipStore())

  const selected = computed(() => data.value)

  function doSelect (element: HTMLElement | null) {
    data.value = element
  }

  return { selected, doSelect }
}
