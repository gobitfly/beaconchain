import type { AgeFormat } from '~/types/settings'

export const useSettingsStore = defineStore('settings', () => {
  const ageFormat = ref<AgeFormat>('absolute')
  const toggleAgeFormat = () => {
    ageFormat.value = ageFormat.value === 'absolute' ? 'relative' : 'absolute'
  }

  const { displayCurrencyDefault } = useNetworkStore()
  const selectedCurrencyMain = ref<CurrencyCode>(displayCurrencyDefault.main)
  const selectedCurrencyExecutionLayer = ref<CurrencyCode>(displayCurrencyDefault.executionLayer)
  const selectedCurrencyConsensusLayer = ref<CurrencyCode>(displayCurrencyDefault.consensusLayer)
  const setCurrencyMain = (currencyCode: CurrencyCode, availableCurrencies: CurrencyCode[]) => {
    if (!availableCurrencies.includes(currencyCode)) {
      logError(`CurrencyCode not found: availableCurrencies does not include ${currencyCode}`)
    }
    selectedCurrencyMain.value = currencyCode
  }

  return {
    ageFormat,
    selectedCurrencyConsensusLayer,
    selectedCurrencyExecutionLayer,
    selectedCurrencyMain,
    setCurrencyMain,
    toggleAgeFormat,
  }
}, {
  persist: true,
})
