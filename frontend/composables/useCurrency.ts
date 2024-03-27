import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { type Currency } from '~/types/currencies'

export function useCurrency () {
  const { latestState } = useLatestStateStore()

  const selectedCurrency = useCookie<Currency>('currency', { default: () => 'NAT' })
  const currency = readonly(selectedCurrency)
  function setCurrency (newCurrency: Currency) {
    selectedCurrency.value = newCurrency
  }

  const rates = computed(() => {
    return latestState.value?.rates || {} as Record<Currency, number>
  })

  const available = computed(() => {
    let list: Currency[] = ['NAT', 'ETH']
    if (latestState.value?.rates) {
      list = list.concat(Object.keys(latestState.value.rates) as Currency[])
    }
    return list
  })

  return { currency, setCurrency, available, rates }
}
