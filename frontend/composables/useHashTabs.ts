import type { HashTabs } from '~/types/hashTabs'

export function useHashTabs(tabs: HashTabs, defaultTab: string, useRouteHash = false) {
  const activeIndex = ref<string>('-1')
  const { hash: initialHash } = useRoute()

  const findFirstValidIndex = () => {
    if (tabs[defaultTab] && !tabs[defaultTab].disabled) {
      return tabs[defaultTab].index
    }
    const list = Object.values(tabs)
    for (let i = 0; i < list.length; i++) {
      const tab = list[i]
      if (!tab.disabled) {
        return tab.index
      }
    }
    return '-1'
  }
  const findHashForIndex = (index: string) => {
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
    const hash = useRouteHash ? initialHash?.replace('#', '') : ''
    activeIndex.value
      = hash && tabs[hash] && !tabs[hash].disabled
        ? tabs[hash].index
        : findFirstValidIndex()
  })

  const updateHash = (index: string) => {
    if (isServerSide || !useRouteHash) {
      return
    }
    window.location.hash = findHashForIndex(index)
  }

  watch(
    activeIndex,
    (index) => {
      if (isServerSide && index === '-1') {
        return
      }
      updateHash(index)
    },
    { immediate: true },
  )

  return {
    activeIndex,
  }
}
