import { reduce } from 'lodash-es'
import { type EthConversionRate } from '~/types/api/latest_state'
import { COOKIE_KEY } from '~/types/cookie'
import { type Currency } from '~/types/currencies'

export function useCurrency () {
  const { latestState } = useLatestStateStore()
  const { networkInfo } = useNetwork()
  const { t: $t } = useI18n()
  const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

  const selectedCurrency = useCookie<Currency>(COOKIE_KEY.CURRENCY, { default: () => 'NAT' })
  function setCurrency (newCurrency: Currency) {
    selectedCurrency.value = newCurrency
  }

  const rates = computed<Partial<Record<Currency, EthConversionRate>>>(() => {
    const rec: Partial<Record<Currency, EthConversionRate>> = {}
    return reduce(
      latestState.value?.exchange_rates || [],
      (list, rate) => {
        list[rate.code as Currency] = rate
        return list
      },
      rec
    )
  })

  const available = computed<Currency[]>(() => {
    const list: Currency[] = [networkInfo.value.elCurrency]
    if (networkInfo.value.clCurrency !== networkInfo.value.elCurrency) {
      list.push(networkInfo.value.clCurrency)
    }
    if (showInDevelopment) {
      list.splice(0, 1, 'NAT')
    }
    return list.concat((latestState.value?.exchange_rates || []).map(r => r.code as Currency))
  })

  const currency = computed(() => selectedCurrency.value && available.value.includes(selectedCurrency.value) ? selectedCurrency.value : available.value[0])

  const withLabel = computed(() => {
    return available.value?.map(currency => ({
      currency,
      label: $t(`currency.label.${currency}`, {}, rates.value?.[currency]?.currency || currency)
    }))
  })

  watch([latestState, selectedCurrency], () => {
    // once we loaded our latestState and see that we don't support the currency we switch back to the first item
    if (latestState.value && !available.value.includes(selectedCurrency.value)) {
      selectedCurrency.value = available.value[0]
    }
  })

  return { currency, setCurrency, available, rates, withLabel }
}
