import { BigNumber } from '@ethersproject/bignumber'
import { formatEther, commify } from '@ethersproject/units'
import type { Currency, CryptoCurrency } from '~/types/currencies'
import type { ValueConvertOptions } from '~/types/value'
import { isNative, isFiat } from '~/utils/currency'
import { OneEther, OneGwei, lessThenGwei, lessThenEth } from '~/utils/ether'

const withCurrency = (value: string, currency: string) => {
  return `${value} ${currency}`
}

const nZeros = (count: number):string => {
  return count > 0 ? Array.from(Array(count)).map(() => '0').join('') : ''
}

const commmifyLeft = (value: string):string => {
  const formatted = commify(value)
  const i = formatted.lastIndexOf('.0')
  if (i >= 0 && i === formatted.length - 2) {
    console.log('commmifyLeft', i, formatted, formatted.substring(0, formatted.length - 2))
    return formatted.substring(0, formatted.length - 2)
  }
  return formatted
}

const trim = (value:string, maxDecimalCount: number, minDecimalCount?: number):string => {
  minDecimalCount = minDecimalCount === undefined ? maxDecimalCount : Math.min(minDecimalCount, maxDecimalCount)
  const split = value.split('.')
  let dec = (split[1] ?? '').substring(0, maxDecimalCount)
  while (dec.length < minDecimalCount) {
    dec += '0'
  }
  if (split[0] === '0' && (!dec || parseInt(dec) === 0)) {
    if (maxDecimalCount === 0) {
      return '<1'
    }
    return `<0.${nZeros(minDecimalCount - 1)}1`
  }
  const left = commmifyLeft(split[0])
  if (!dec?.length) {
    return left
  }
  return `${left}.${dec}`
}

export function useValue () {
  const { currency, rates } = useCurrency()

  interface ValueResult {
    label: string,
    fullLabel?: string;
  }

  const converter = computed(() => {
    const weiToValue = (wei?:string | BigNumber, options?: ValueConvertOptions):ValueResult => {
      if (!wei) {
        return { label: '' }
      }
      const useNative = isNative(currency.value)
      const source: CryptoCurrency = options?.sourceCurrency ?? 'ETH'
      const target: Currency = options?.targetCurrency || useNative ? source : currency.value
      let value = typeof wei === 'string' ? BigNumber.from(wei) : wei
      if (value.isZero()) {
        return {
          label: withCurrency('0', target)
        }
      }
      // If we don't show the native currency but the input is not ETH we first convert it to ETH
      if (!useNative && source !== 'ETH' && source !== target && rates.value[source]) {
        value = value.mul(rates.value[source])
      }
      // If source and target are different we convert it in the target currency
      if (source !== target && rates.value[target]) {
        value = value.div(rates.value[source])
      }
      let currencyLabel:string = target
      const minDecimalCount: number | undefined = options?.fixedDecimalCount
      let maxDecimalCount: number = options?.fixedDecimalCount ?? 5

      if (isFiat(target)) {
        maxDecimalCount = Math.min(maxDecimalCount, 2)
      }
      if (useNative && (!options?.minUnit || options?.minUnit !== 'MAIN')) {
        if ((!options?.minUnit || options?.minUnit === 'WEI') && lessThenGwei(value, maxDecimalCount)) {
          value = value.mul(OneEther)
          maxDecimalCount = 0
          currencyLabel = 'WEI'
        } else if ((!options?.minUnit || options?.minUnit === 'GWEI') && lessThenEth(value, maxDecimalCount)) {
          value = value.mul(OneGwei)
          currencyLabel = 'GWEI'
        }
      }

      const ether = formatEther(value)
      const label = trim(ether, maxDecimalCount, minDecimalCount)
      const fullLabel = commmifyLeft(ether)

      return {
        label: withCurrency(label, currencyLabel),
        fullLabel: (label.startsWith('<') || commify(label.replaceAll(',', '')) !== fullLabel) ? withCurrency(fullLabel, currencyLabel) : undefined
      }
    }
    return { weiToValue }
  })

  return { converter }
}
