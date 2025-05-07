export type DateTimeFormat = 'absolute' | 'relative'

export const useSettingsStore = defineStore('settings', () => {
  const dateTimeFormat = ref<DateTimeFormat>('relative')
  const toggleDateTimeFormat = () => {
    dateTimeFormat.value = dateTimeFormat.value === 'absolute' ? 'relative' : 'absolute'
  }

  const { displayCurrencyDefault } = useNetwork()
  const selectedCurrencyMain = ref<CurrencyCode>(displayCurrencyDefault.main)

  return {
    dateTimeFormat,
    selectedCurrencyMain,
    toggleDateTimeFormat,
  }
}, {
  persist: true,
})
