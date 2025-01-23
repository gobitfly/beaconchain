<script setup lang="ts" generic="UNUSED_WORKAROUND_FOR_CONDITIONAL_PROPS">
// https://github.com/vuejs/core/issues/8952

type CurrencyItem = {
  consensusLayerValue: number | string,
  executionLayerValue: number | string,
} | {
  consensusLayerValue: number | string,
  executionLayerValue?: never,

} | {
  consensusLayerValue?: never,
  executionLayerValue: number | string,
}

type FormatAmountOptions = (
    {
      currencyItems: CurrencyItem[],
      sourceCurrency?: never,
      value?: never,
    }
  |
    {
      currencyItems?: never,
      sourceCurrency?: CurrencyCode,
      value: `${number}` | string,
    }
) & {
  fractionDigits?: number,
  hasColor?: boolean,
  hasDashForZero?: boolean,
  hasHigherPrecision?: boolean,
  hasSignDisplay?: boolean,
  hasTooltip?: boolean,
  targetCurrency?: CurrencyCode,
  targetUnitCrypto?: 'auto',
  zeroDisplay?: 'auto' | 'dash',
}

const {
  zeroDisplay = 'dash',
  ...props
} = defineProps<FormatAmountOptions>()

const {
  addCurrencies,
  clCurrency,
  elCurrency,
  formatAmount,
  fractionDigitsDefault,
  selectedCurrencyMain,
} = useCurrency()

const currencyItems = computed(() => {
  const result: { sourceCurrency: CurrencyCode, value: number | string }[] = []
  props.currencyItems?.forEach((item) => {
    if (item.executionLayerValue) {
      result.push({
        sourceCurrency: elCurrency,
        value: item.executionLayerValue,
      })
    }
    if (item.consensusLayerValue) {
      result.push({
        sourceCurrency: clCurrency,
        value: item.consensusLayerValue,
      })
    }
  })
  return result
})

const amount = computed(() => {
  if (props.value) return `${props.value}`
  return addCurrencies({
    currencyItems: currencyItems.value,
  })
})

const color = computed(() => {
  if (props.hasColor) {
    if (Number(amount.value) > 0) return 'green'
    if (Number(amount.value) < 0) return 'red'
  }
  return undefined
})
const signDisplay = computed(() => {
  if (props.hasSignDisplay) return 'exceptZero'
  return undefined
})

const formattedAmount = computed(() => {
  return formatAmount(amount.value, {
    hasHigherPrecision: props.hasHigherPrecision,
    hasUnitDisplay: !!props.targetUnitCrypto,
    signDisplay: signDisplay.value,
    sourceCurrency: props.sourceCurrency,
    targetCurrency: props.targetCurrency,
  })
})

// we want to show values in Gwei to avoid `0.000000 GNO`
const isCryptoAmountTooSmall = computed(() => {
  const [
    value,
    _unit,
  ] = formattedAmount.value.split(' ')
  const fractionDigits = fractionDigitsDefault.crypto.base
  if (isFiat(selectedCurrencyMain.value)) return false
  if (amount.value === '0') return false
  if (Number(value) > (10 ** -fractionDigits)) return false
  return true
})

const formattedAmountWithUnit = computed(() => {
  const shouldShowInGwei = !!(props.targetUnitCrypto === 'auto' && isCryptoAmountTooSmall.value)
  // console.log('👉', props.value, amount.value, formatAmount(amount.value, {
  //   hasHigherPrecision: props.hasHigherPrecision,
  //   hasUnitDisplay: shouldShowInGwei,
  //   maximumFractionDigits: props.fractionDigits,
  //   minimumFractionDigits: props.fractionDigits,
  //   signDisplay: signDisplay.value,
  //   sourceCurrency: props.sourceCurrency,
  //   targetCurrency: props.targetCurrency,
  //   targetUnit: shouldShowInGwei ? 'gwei' : undefined,
  // }))
  return formatAmount(amount.value, {
    hasHigherPrecision: props.hasHigherPrecision,
    hasUnitDisplay: shouldShowInGwei,
    maximumFractionDigits: props.fractionDigits,
    minimumFractionDigits: props.fractionDigits,
    signDisplay: signDisplay.value,
    sourceCurrency: props.sourceCurrency,
    targetCurrency: props.targetCurrency,
    targetUnit: shouldShowInGwei ? 'gwei' : undefined,
  })
})
const formattedAmountWithUnitHigherPrecision = computed(() => {
  const shouldShowInGwei = !!(props.targetUnitCrypto === 'auto' && isCryptoAmountTooSmall.value)
  return formatAmount(amount.value, {
    hasHigherPrecision: true,
    hasUnitDisplay: shouldShowInGwei,
    maximumFractionDigits: props.fractionDigits,
    minimumFractionDigits: props.fractionDigits,
    signDisplay: signDisplay.value,
    sourceCurrency: props.sourceCurrency,
    targetCurrency: props.targetCurrency,
    targetUnit: shouldShowInGwei ? 'gwei' : undefined,
  })
})
</script>

<template>
  <span v-if="amount === '0' && zeroDisplay === 'dash'">
    -
  </span>
  <BcTooltip
    v-else
    fit-content
  >
    <template
      v-if="hasTooltip"
      #tooltip
    >
      {{ formattedAmountWithUnitHigherPrecision }}
    </template>
    <span
      :class="color"
    >
      <slot :value="formattedAmountWithUnit">
        {{ formattedAmountWithUnit }}
      </slot>
    </span>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.green {
  color: var(--positive-color);
}
.red {
  color: var(--negative-color);
}
</style>
