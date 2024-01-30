import { defineStore } from 'pinia'

export const useTooltipStore = defineStore('user-store', () => {
  const selected = ref<HTMLElement | null>(null)

  function doSelect (element: HTMLElement | null) {
    selected.value = element
  }

  return { selected, doSelect }
})
