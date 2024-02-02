import { storeToRefs } from 'pinia'
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { type Currency } from '~/types/currencies'

export function useCurrency () {
  const { latest } = storeToRefs(useLatestStateStore())

  const selectedCurrency = useCookie<Currency>('currency', { default: () => 'NAT' })
  const currency = readonly(selectedCurrency)
  function setCurrency (newCurrency: Currency) {
    selectedCurrency.value = newCurrency
  }

  const rates = computed(() => {
    return latest.value?.rates || {} as Record<Currency, number>
  })

  const available = computed(() => {
    let list: Currency[] = ['NAT', 'ETH']
    if (latest.value?.rates) {
      list = list.concat(Object.keys(latest.value.rates) as Currency[])
    }
    return list
  })

  return { currency, setCurrency, available, rates }
}
