import type { AgeFormat } from '~/types/settings'

export const useSettingsStore = defineStore('settings', () => {
  const ageFormat = ref<AgeFormat>('absolute')
  const toggleAgeFormat = () => {
    ageFormat.value = ageFormat.value === 'absolute' ? 'relative' : 'absolute'
  }

  const { displayCurrencyDefault } = useNetworkStore()
  const selectedCurrencyMain = ref<CurrencyCode>(displayCurrencyDefault.main)

  return {
    ageFormat,
    selectedCurrencyMain,
    toggleAgeFormat,
  }
}, {
  persist: true,
})
