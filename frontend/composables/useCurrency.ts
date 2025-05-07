export const useCurrency = () => {
  const {
    clCurrency,
    displayCurrencyDefault,
    elCurrency,
  } = useNetwork()

  const settingsStore = useSettingsStore()
  const {
    selectedCurrencyMain,
  } = storeToRefs(settingsStore)

  const { latestState } = storeToRefs(useLatestStateStore())
  const exchangeRates = computed(() => latestState.value?.exchange_rates ?? [])

  const convertCurrency = ({
    sourceCurrency,
    targetCurrency,
    value,
  }: {
    sourceCurrency: CurrencyCode,
    targetCurrency: CurrencyCode,
    value: number | string,
  }) => {
    assertIsNumber(value)
    if (sourceCurrency === targetCurrency) {
      return `${value}`
    }

    const getExchangeRate = (currencyCode: CurrencyCode) => {
      const conversionInfo = exchangeRates.value.find(exchangeRate => exchangeRate.code === currencyCode)
      if (!conversionInfo) {
        logError(`Currency not found in exchangeRates. Missing currency is ${currencyCode}`)
        return 1
      }
      return conversionInfo.rate
    }

    const exchangeRateSourceCurrency = `${getExchangeRate(sourceCurrency)}`
    const valueInMainCurrency = divide(value, exchangeRateSourceCurrency)

    const exchangeRateTargetCurrency = `${getExchangeRate(targetCurrency)}`
    const valueInTargetCurrency = multiply(
      valueInMainCurrency,
      exchangeRateTargetCurrency,
    )
    return `${valueInTargetCurrency}`
  }

  const fractionDigitsDefault = {
    crypto: {
      base: 6,
      highPrecision: 18,
    },
    fiat: {
      base: 2,
      highPrecision: 4,
    },
  } as const

  const getFractionDigitDefault = ({
    hasHigherPrecision,
    targetCurrency,
  }: {
    hasHigherPrecision: boolean,
    targetCurrency: CurrencyCode,
  }) => {
    if (isFiat(targetCurrency)) {
      return hasHigherPrecision
        ? fractionDigitsDefault.fiat.highPrecision
        : fractionDigitsDefault.fiat.base
    }

    return hasHigherPrecision
      ? fractionDigitsDefault.crypto.highPrecision
      : fractionDigitsDefault.crypto.base
  }

  const { t: $t } = useTranslation()
  const getCurrencyName = (currencyCode: CurrencyCode) => {
    if (currencyCode === 'AUD') return $t('currency.label.AUD')
    if (currencyCode === 'CAD') return $t('currency.label.CAD')
    if (currencyCode === 'CNY') return $t('currency.label.CNY')
    if (currencyCode === 'EUR') return $t('currency.label.EUR')
    if (currencyCode === 'GBP') return $t('currency.label.GBP')
    if (currencyCode === 'JPY') return $t('currency.label.JPY')
    if (currencyCode === 'USD') return $t('currency.label.USD')

    if (currencyCode === 'ETH') return $t('currency.label.ETH')
    if (currencyCode === 'GNO') return $t('currency.label.GNO')
    if (currencyCode === 'mGNO') return $t('currency.label.mGNO')
    if (currencyCode === 'DAI') return $t('currency.label.DAI')
    if (currencyCode === 'xDAI') return $t('currency.label.xDAI')
  }
  /**
   *
   * @param {CurrencyCode} [options.sourceCurrency] - Default: clCurrency
   * @param {CurrencyCode} [options.targetCurrency] - Default: selectedCurrencyMain
   * @param {CurrencyCode} [options.sourceUnit] - Default: wei
   * @param {CurrencyCode} [options.targetUnit] - Default: base
   * @param {boolean} [options.hasHigherPrecision] - Default: false
   * Whether to use more fractions as defaults for precision for
   * maximumFractionDigits and minimumFractionDigits depending on targetCurrency
   * @param {number} [options.maximumFractionDigits] - Defaults:
   *   - fiat: 2 (4 if hasHigherPrecision=true)
   *   - crypto: 6 (18 if hasHigherPrecision=true)
   * @returns {string} The formatted amount with currency code
   */
  const formatAmount = (
    value: string,
    options: {
      hasCurrencyDisplay?: boolean,
      hasHigherPrecision?: boolean,
      hasRoundingIndication?: boolean,
      hasUnitDisplay?: boolean,
      maximumFractionDigits?: number,
      minimumFractionDigits?: number,
      signDisplay?: Intl.NumberFormatOptions['signDisplay'],
      sourceCurrency?: CurrencyCode,
      sourceUnit?: CryptoUnit,
      targetCurrency?: CurrencyCode,
      targetUnit?: CryptoUnit,
      useGrouping?: Intl.NumberFormatOptions['useGrouping'],
    } = {},
  ): string => {
    assertIsNumber(value)
    const {
      hasCurrencyDisplay = true,
      hasHigherPrecision = false,
      hasRoundingIndication,
      hasUnitDisplay = false,
      targetCurrency = selectedCurrencyMain.value,
      maximumFractionDigits = getFractionDigitDefault({
        hasHigherPrecision,
        targetCurrency,
      }),
      minimumFractionDigits = isFiat(targetCurrency) ? 2 : undefined,
      signDisplay,
      sourceCurrency = clCurrency,
      sourceUnit = 'wei',
      targetUnit = 'base',
      useGrouping,
    } = options

    const valueConverted = convertCurrency({
      sourceCurrency,
      targetCurrency,
      value,
    })

    const unitFactor = unitFactorCrypto[sourceUnit] - unitFactorCrypto[targetUnit]

    const unitTranslation = {
      gwei: $t('common.units.gwei'),
      wei: $t('common.units.wei'),
    }
    const unit = hasUnitDisplay && targetUnit !== 'base' ? ` (${unitTranslation[targetUnit]})` : ''
    const formattedValue = `${formatNumber(valueConverted, {
      hasRoundingIndication,
      maximumFractionDigits,
      minimumFractionDigits,
      scaleBy: unitFactor,
      signDisplay,
      useGrouping,
    })}`
    const currency = hasCurrencyDisplay ? ` ${targetCurrency}` : ''
    return `${formattedValue}${unit}${currency}`
  }

  /**
   * Sums up an array of objects that contains amounts of different currencies.
   *
   * @param {Array} options.currencyItems - The array of currency items to sum up.
   * @param {CurrencyCode} [options.currencyItems[].sourceCurrency=clCurrency] - Default: clCurrency
   * @param {number | string} options.currencyItems[].value - The value of the currency item.
   * @param {CurrencyCode} [options.targetCurrency=clCurrency] - Default: clCurrency
   * @returns {string} The sum of the currency values converted to the target currency.
   */
  const addCurrencies = (
    {
      currencyItems,
      targetCurrency = clCurrency,
    }:
    {
      currencyItems: {
        sourceCurrency?: CurrencyCode,
        value: number | string,
      }[],
      targetCurrency?: CurrencyCode,
    },
  ) => {
    return currencyItems.reduce((sum, currencyItem) => {
      const {
        sourceCurrency = clCurrency,
        value,
      } = currencyItem
      const convertedCurrencyValue = convertCurrency({
        sourceCurrency,
        targetCurrency,
        value,
      })
      return sum = `${add(sum, convertedCurrencyValue)}`
    }, '0')
  }

  return {
    addCurrencies,
    clCurrency,
    displayCurrencyDefault,
    elCurrency,
    exchangeRates,
    formatAmount,
    fractionDigitsDefault,
    getCurrencyName,
    selectedCurrencyMain,
  }
}
