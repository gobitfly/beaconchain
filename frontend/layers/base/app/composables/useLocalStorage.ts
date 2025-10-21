import { useStorage } from '@vueuse/core'

type LocalStorageKey = 'bc-search-history-product-landing'

export const useLocalStorage
  = <T extends MaybeRefOrGetter<boolean | null | number | Record<PropertyKey, any> | string>>
  (key: LocalStorageKey, value: T) => {
    const state = useStorage<T>(key, value)
    return state
  }
