export const useExchangeRates = () => {
  const latestState = useFetchedData('/api/latest-state')
  const exchangeRates = computed(() => latestState.value?.exchange_rates ?? [])
  return exchangeRates
}
