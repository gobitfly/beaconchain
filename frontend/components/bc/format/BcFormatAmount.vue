<script setup lang="ts" generic="WORKAROUND_FOR_CONDITIONAL_PROPS">
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
      sourceCurrency?: 'clCurrency' | 'elCurrency' | CurrencyCode,
      value: `${number}` | string,
    }
) & {
  hasAdditionalSelectedCurrencyMain?: boolean,
  hasColor?: boolean,
  hasDashForZero?: boolean,
  hasHigherPrecision?: boolean,
  hasSignDisplay?: boolean,
  hasTooltip?: boolean,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  /**
   * @description
   * Display currencies take into account, that the currency in the binary data
   * can be different to what should be displayed to the user.
   *
   * E.g.: we get mGno values but want to display them in GNO
   */
  targetCurrency?: 'clDisplayCurrency' | 'elDisplayCurrency' | CurrencyCode,
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
  displayCurrencyDefault,
  elCurrency,
  formatAmount,
  fractionDigitsDefault,
  selectedCurrencyMain,
} = useCurrency()

const sourceCurrency = computed(() => {
  if (props.sourceCurrency === 'elCurrency') return elCurrency
  if (props.sourceCurrency === 'clCurrency') return clCurrency
  return props.sourceCurrency
})

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

const targetCurrency = computed(() => {
  if (props.targetCurrency === 'elDisplayCurrency') return displayCurrencyDefault.executionLayer
  if (props.targetCurrency === 'clDisplayCurrency') return displayCurrencyDefault.consensusLayer
  if (!props.targetCurrency) return selectedCurrencyMain.value
  return props.targetCurrency
})

const formattedAmount = computed(() => {
  return formatAmount(amount.value, {
    hasCurrencyDisplay: false,
    hasUnitDisplay: false,
    sourceCurrency: sourceCurrency.value,
    targetCurrency: targetCurrency.value,
  })
})

const format = (value: string, optionsOverride?: Parameters<typeof formatAmount>[1]) => {
  const getTargetUnit = () => {
    const fractionDigits = fractionDigitsDefault.crypto.base
    if (isFiat(optionsOverride?.targetCurrency ?? targetCurrency.value)) return 'base'
    if (props.targetUnitCrypto !== 'auto') return 'base'
    if (amount.value === '0') return 'base'
    if (Math.abs(Number(formattedAmount.value)) > (10 ** -fractionDigits)) return 'base'
    return 'gwei'
  }
  return formatAmount(value, {
    hasHigherPrecision: props.hasHigherPrecision,
    hasUnitDisplay: getTargetUnit() !== 'base',
    maximumFractionDigits: props.maximumFractionDigits
      ?? (
        getTargetUnit() !== 'base'
          ? fractionDigitsDefault.crypto.base
          : undefined
      ),
    minimumFractionDigits: props.minimumFractionDigits,
    signDisplay: signDisplay.value,
    sourceCurrency: sourceCurrency.value,
    targetCurrency: targetCurrency.value,
    targetUnit: getTargetUnit(),
    ...optionsOverride,
  })
}
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
      {{ format(amount, {
        hasHigherPrecision: true,
      }) }}
    </template>
    <span
      :class="color"
    >
      <slot :value="format(amount)">
        <span>
          {{ format(amount) }}
        </span>
        <span
          v-if="hasAdditionalSelectedCurrencyMain && targetCurrency !== selectedCurrencyMain"
        >
          ({{ format(amount, {
            targetCurrency: selectedCurrencyMain,
          }) }})
        </span>
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
