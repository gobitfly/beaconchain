<script setup lang="ts">
import {
  LazyBcIconCryptoEth,
  LazyBcIconCryptoGno,
  LazyBcIconCryptoXdai,
} from '#components'

const { currencyCode } = defineProps<{
  colorMode?: 'colored' | 'currentColor' | 'dark' | 'light',
  currencyCode: CurrencyCodeCrypto,
}>()

const is = computed(() => {
  if (currencyCode === 'xDAI' || currencyCode === 'DAI') return LazyBcIconCryptoXdai
  if (currencyCode === 'GNO' || currencyCode === 'mGNO') return LazyBcIconCryptoGno
  if (currencyCode === 'ETH') return LazyBcIconCryptoEth
  return 'span'
})
</script>

<template>
  <Component
    :is
    :color-mode
  />
</template>
