import type { HashTabs } from '~/types/hashTabs'

export function useHashTabs(tabs: HashTabs) {
  const activeIndex = ref(-1)
  const { hash: initialHash } = useRoute()

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
      const [
        hash,
        tab,
      ] = entries[i]
      if (!tab.disabled && tab.index === index) {
        return `#${hash}`
      }
    }
    return ''
  }

  onMounted(() => {
    const hash = initialHash?.replace('#', '')
    activeIndex.value
      = hash && tabs[hash] && !tabs[hash].disabled
        ? tabs[hash].index
        : findFirstValidIndex()
  })

  const updateHash = (index: number) => {
    if (isServerSide) {
      return
    }
    window.location.hash = findHashForIndex(index)
  }

  watch(
    activeIndex,
    (index) => {
      if (isServerSide && index < 0) {
        return
      }
      updateHash(index)
    },
    { immediate: true },
  )

  const setActiveIndex = (index: number) => {
    if (isServerSide) {
      return
    }
    activeIndex.value = index
    updateHash(index)
  }

  return {
    activeIndex,
    setActiveIndex,
  }
}
