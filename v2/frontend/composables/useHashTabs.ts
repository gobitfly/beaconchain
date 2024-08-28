import type { HashTabs } from '~/types/hashTabs'

export function useHashTabs(tabs: HashTabs, defaultTab: string, useRouteHash = false) {
  const activeTab = ref<string>('-1')
  const { hash: initialHash } = useRoute()

  const findFirstValidIndex = () => {
    const defaultKey = tabs.find(t => t.key === defaultTab)
    if (defaultKey) {
      return defaultKey.key
    }

    for (let i = 0; i < tabs.length; i++) {
      const tab = tabs[i]
      if (!tab.disabled) {
        return tab.key
      }
    }
    return '-1'
  }

  onMounted(() => {
    const hash = useRouteHash ? initialHash?.replace('#', '') : ''
    const matchedTab = tabs.find(t => t.key === hash)
    activeTab.value
      = hash && matchedTab && !matchedTab.disabled
        ? matchedTab.key
        : findFirstValidIndex()
  })

  const updateHash = (key: string) => {
    if (isServerSide || !useRouteHash) {
      return
    }
    window.location.hash = key
  }

  watch(
    activeTab,
    (key) => {
      if (isServerSide || key === '-1') {
        return
      }
      updateHash(key)
    },
    { immediate: true },
  )

  return {
    activeTab,
  }
}
