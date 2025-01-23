<script setup lang="ts">
import {
  BcIconCryptoEth,
  BcIconCryptoGno,
  BcIconCryptoXdai,
} from '#components'

const { currencyCode } = defineProps<{
  colorMode?: 'colored' | 'currentColor' | 'dark' | 'light',
  currencyCode: CurrencyCodeCrypto,
}>()

const is = computed(() => {
  if (currencyCode === 'xDAI' || currencyCode === 'DAI') return BcIconCryptoXdai
  if (currencyCode === 'GNO' || currencyCode === 'mGNO') return BcIconCryptoGno
  if (currencyCode === 'ETH') return BcIconCryptoEth
  return 'span'
})
</script>

<template>
  <Component
    :is
    :color-mode
  />
</template>
