import type { HashTabs } from '~/types/hashTabs'

export function useHashTabs (tabs: HashTabs) {
  const route = useRoute()
  const activeIndex = ref(-1)

  const findFirstValidIndex = () => {
    const list = Object.values(tabs)
    for (let i = 0; i < list.length; i++) {
      const tab = list[i]
      if (!tab.disabled) {
        return tab.index
      }
    }
    return -1
  }
  const findHashForIndex = (index: number) => {
    const entries = Object.entries(tabs)
    for (let i = 0; i < entries.length; i++) {
      const [hash, tab] = entries[i]
      if (!tab.disabled && tab.index === index) {
        return hash
      }
    }
    return ''
  }

  onMounted(() => {
    const hash = route.hash.replace('#', '')
    activeIndex.value = hash && tabs[hash] && !tabs[hash].disabled ? tabs[hash].index : findFirstValidIndex()
  })

  watch(activeIndex, (index) => {
    if (process.server) {
      return
    }
    window.location.hash = findHashForIndex(index)
  }, { immediate: true })

  const setActiveIndex = (index: number) => {
    if (process.server) {
      return
    }
    activeIndex.value = index
    window.location.hash = findHashForIndex(index)
  }

  return { activeIndex, setActiveIndex }
}
