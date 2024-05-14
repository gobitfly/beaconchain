import { BigNumber } from '@ethersproject/bignumber'
import { formatEther, commify } from '@ethersproject/units'
import type { Currency, CryptoCurrency } from '~/types/currencies'
import type { ExtendedLabel, WeiToValue, ValueConvertOptions } from '~/types/value'
import { isNative, isFiat } from '~/utils/currency'
import { OneEther, OneGwei, lessThanGwei, lessThanEth } from '~/utils/ether'
import { commmifyLeft, trim, withCurrency } from '~/utils/format'

export function useValue () {
  const { currency, rates } = useCurrency()

  const converter = computed(() => {
    const weiToValue:WeiToValue = (wei?: string | BigNumber, options?: ValueConvertOptions): ExtendedLabel => {
      if (!wei) {
        return { label: '' }
      }
      const useNative = isNative(currency.value)
      const source: CryptoCurrency = options?.sourceCurrency ?? 'ETH'
      const target: Currency = options?.targetCurrency || (useNative ? source : currency.value)
      let value = typeof wei === 'string' ? BigNumber.from(wei) : wei
      if (value.isZero()) {
        if (options?.fixedDecimalCount) {
          return {
            label: withCurrency(`0.${nZeros(options?.fixedDecimalCount)}`, target)
          }
        }
        return {
          label: withCurrency('0', target)
        }
      }

      // If a different sourceUnit is defined we multiply accordingly. We usually get gwei, but you never know.
      if (options?.sourceUnit === 'GWEI') {
        value = value.mul(OneGwei)
      } else if (options?.sourceUnit === 'MAIN') {
        value = value.mul(OneEther)
      }

      // If we don't show the native currency but the input is not ETH we first convert it to ETH
      if (!useNative && source !== 'ETH' && source !== target && rates.value[source]?.rate) {
        value = bigDiv(value, rates.value[source]!.rate)
      }
      // If source and target are different we convert it in the target currency
      if (source !== target && rates.value[target]?.rate) {
        value = bigMul(value, rates.value[target]!.rate)
      }
      let currencyLabel: string = target
      const minDecimalCount: number | undefined = options?.fixedDecimalCount
      let maxDecimalCount: number = options?.fixedDecimalCount ?? 5

      if (isFiat(target)) {
        maxDecimalCount = Math.min(maxDecimalCount, 2)
      }
      if (useNative && (!options?.minUnit || options?.minUnit !== 'MAIN')) {
        if (options?.fixedUnit === 'WEI' || ((!options?.minUnit || options?.minUnit === 'WEI') && lessThanGwei(value.abs(), options?.minUnitDecimalCount ?? maxDecimalCount))) {
          value = value.mul(OneEther)
          maxDecimalCount = 0
          currencyLabel = 'WEI'
        } else if (options?.fixedUnit === 'GWEI' || ((!options?.minUnit || options?.minUnit === 'GWEI') && lessThanEth(value.abs(), options?.minUnitDecimalCount ?? maxDecimalCount))) {
          value = value.mul(OneGwei)
          currencyLabel = 'GWEI'
        }
      }

      const ether = formatEther(value)
      const label = trim(ether, maxDecimalCount, minDecimalCount)
      const fullLabel = commmifyLeft(ether)
      const fullRequired = (label.startsWith('<') || !commify(label.replaceAll(',', '')).startsWith(fullLabel))

      return {
        label: withCurrency(addPlusSign(label, options?.addPlus === true), currencyLabel),
        fullLabel: fullRequired ? withCurrency(addPlusSign(fullLabel, options?.addPlus === true), currencyLabel) : undefined
      }
    }
    return { weiToValue }
  })

  return { converter }
}
