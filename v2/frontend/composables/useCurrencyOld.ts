import { reduce } from 'lodash-es'
import type { EthConversionRate } from '~/types/api/latest_state'

export function useCurrencyOld() {
  const latestStateStore = useLatestStateStore()
  const { latestState } = storeToRefs(latestStateStore)
  const { networkInfo } = useNetworkStore()
  const { t: $t } = useTranslation()
  const available = ref<CurrencyCode[]>([])
  const withLabel = ref<{ currency: CurrencyCode,
    label: string, }[]>([])

  const rates = computed<Partial<Record<CurrencyCode, EthConversionRate>>>(() => {
    const rec: Partial<Record<CurrencyCode, EthConversionRate>> = {}
    return reduce(
      latestState.value?.exchange_rates || [],
      (list, rate) => {
        list[rate.code as CurrencyCode] = rate
        return list
      },
      rec,
    )
  })

  watch(
    [
      latestState,
      networkInfo,
    ],
    () => {
      let list: CurrencyCode[] = [ networkInfo.value.elCurrency ]
      if (networkInfo.value.clCurrency !== networkInfo.value.elCurrency) {
        list.push(networkInfo.value.clCurrency)
      }
      list = list.concat(
        (latestState.value?.exchange_rates || []).map(
          r => r.code as CurrencyCode,
        ),
      )
      // make sure we update the currency list only if it really changed (to prevent reactivity triggers)
      if (JSON.stringify(list) !== JSON.stringify(available.value)) {
        available.value = list

        withLabel.value = list.map(currency => ({
          currency,
          label: $t(
            `currency.label.${currency}`,
            {},
            rates.value?.[currency]?.currency || currency,
          ),
        }))
      }
    },
    { immediate: true },
  )

  const currency = computed(() =>
    // selectedCurrency.value && available.value.includes(selectedCurrency.value)
    //   ? selectedCurrency.value
    //   : available.value[0],
    'ETH',
  )

  // watch([
  //   latestState,
  //   selectedCurrency,
  // ], () => {
  //   // once we loaded our latestState and see that we don't support the currency we switch back to the first item
  //   if (
  //     latestState.value
  //     && !available.value.includes(selectedCurrency.value)
  //   ) {
  //     selectedCurrency.value = available.value[0]
  //   }
  // })

  return {
    available,
    currency,
    rates,
    withLabel,
  }
}
