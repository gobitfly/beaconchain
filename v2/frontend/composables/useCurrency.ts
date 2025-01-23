export const useCurrency = (
  defaults?: {
    maximumFractionDigits?: number,
    minimumFractionDigits?: number,
  }) => {
  const {
    clCurrency,
    displayCurrencyDefault,
    elCurrency,
  } = useNetworkStore()

  const settingsStore = useSettingsStore()
  const {
    selectedCurrencyMain,
  } = storeToRefs(settingsStore)
  // const { setCurrencyMain: setCurrencyMainFromSettings } = settingsStore

  const { latestState } = storeToRefs(useLatestStateStore())
  const exchangeRates = computed(() => latestState.value?.exchange_rates ?? [])

  const availableCurrencies = computed(() => exchangeRates.value.map(exchangeRate => exchangeRate.code as CurrencyCode))
  // const setCurrencyMain = (currencyCode: CurrencyCode) => {
  //   setCurrencyMainFromSettings(currencyCode, availableCurrencies.value)
  // }

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

    const exchangeRateoMainCurrencyToSourceCurrency = `${getExchangeRate(sourceCurrency)}`
    const valueInMainCurrency = divideBigNumbers(value, exchangeRateoMainCurrencyToSourceCurrency)

    const exchangeRateMainCurrencyToTargetCurrency = `${getExchangeRate(targetCurrency)}`
    const valueInTargetCurrency = multiplyBigNumbers(
      valueInMainCurrency,
      exchangeRateMainCurrencyToTargetCurrency,
    )
    return `${valueInTargetCurrency}`
  }

  const fractionDigitsDefault = {
    crypto: {
      base: 6,
      heighPrecision: 8,
    },
    fiat: {
      base: 2,
      heighPrecision: 4,
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
        ? fractionDigitsDefault.fiat.heighPrecision
        : fractionDigitsDefault.fiat.base
    }
    return hasHigherPrecision
      ? fractionDigitsDefault.crypto.heighPrecision
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
   * @param {CurrencyCode} [options.sourceCurrency] - (Default: clCurrency)
   * @param {CurrencyCode} [options.targetCurrency] - (Default: selectedCurrencyMain)
   * @param {CurrencyCode} [options.sourceUnit] - (Default: wei)
   * @param {CurrencyCode} [options.targetUnit] - (Default: base)
   * @param {boolean} [options.hasHigherPrecision] - (Default: false)
   * Whether to use more fractions as defaults for precision for
   * maximumFractionDigits and minimumFractionDigits depending on targetCurrency
   * @param {number} [options.maximumFractionDigits] - (Default: 6/8 and 2/4) The maximum number of fraction digits.
   * Defaults to 2 (4 if hasHigherPrecision=true) fiat and 6 (8 if hasHigherPrecision=true) for crypto.
   * @param {number} [options.minimumFractionDigits] - (Default: 6/8 and 2/4) The minimum number of fraction digits.
   * Defaults to 2  (4 if hasHigherPrecision=true) fiat and 6 (8 if hasHigherPrecision=true) for crypto.
   * @returns {string} The formatted amount with currency code.
   */
  const formatAmount = (
    value: string,
    options: {
      hasCurrencyDisplay?: boolean,
      hasHigherPrecision?: boolean,
      hasUnitDisplay?: boolean,
      maximumFractionDigits?: number,
      minimumFractionDigits?: number,
      signDisplay?: Intl.NumberFormatOptions['signDisplay'],
      sourceCurrency?: CurrencyCode,
      sourceUnit?: CryptoUnit,
      targetCurrency?: CurrencyCode,
      targetUnit?: CryptoUnit,
    } = {},
  ): string => {
    assertIsNumber(value)
    const optionsWithDefaults = {
      ...defaults,
      ...options,
    }
    const {
      hasCurrencyDisplay = true,
      hasHigherPrecision = false,
      hasUnitDisplay = false,
      signDisplay,
      sourceCurrency = clCurrency,
      sourceUnit = 'wei',
      targetCurrency = selectedCurrencyMain.value,
      targetUnit = 'base',
    } = optionsWithDefaults
    const {
      maximumFractionDigits = getFractionDigitDefault({
        hasHigherPrecision,
        targetCurrency,
      }),
      minimumFractionDigits = getFractionDigitDefault({
        hasHigherPrecision,
        targetCurrency,
      }),
    } = optionsWithDefaults

    const valueConverted = convertCurrency({
      sourceCurrency,
      targetCurrency,
      value,
    })

    const unitFactor = unitFactorCrypto[sourceUnit] - unitFactorCrypto[targetUnit]

    const unitTranslation = {
      base: '',
      gwei: $t('common.units.gwei'),
      wei: $t('common.units.wei'),
    }
    const unit = hasUnitDisplay && targetUnit !== 'base' ? ` (${unitTranslation[targetUnit]})` : ''
    const currency = hasCurrencyDisplay ? ` ${targetCurrency}` : ''
    const formattedValue = `${formatNumber(valueConverted, {
      maximumFractionDigits,
      minimumFractionDigits,
      scaleBy: unitFactor,
      signDisplay,
    })}`
    return `${formattedValue}${unit}${currency}`
  }

  /**
   * Sums up an array of objects that contain amounts of different currencies.
   *
   * @param {Array} options.currencyItems - The array of currency items to sum up.
   * @param {CurrencyCode} [options.currencyItems[].sourceCurrency=clCurrency] (Default: clCurrency)
   * @param {number | string} options.currencyItems[].value - The value of the currency item.
   * @param {CurrencyCode} [options.targetCurrency=clCurrency] (Default: clCurrency)
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
      return sum = `${addBigNumbers(sum, convertedCurrencyValue)}`
    }, '0')
  }

  return {
    addCurrencies,
    availableCurrencies,
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
