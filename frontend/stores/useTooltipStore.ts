import { defineStore } from 'pinia'

const tooltipStore = defineStore('tooltip_store', () => {
  const data = ref<HTMLElement | null>(null)
  return { data }
})

export function useTooltipStore() {
  const { data } = storeToRefs(tooltipStore())

  const selected = computed(() => data?.value)

  function doSelect(element: HTMLElement | null) {
    data.value = element
  }

  return {
    doSelect,
    selected,
  }
}
