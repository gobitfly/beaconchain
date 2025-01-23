<script setup lang="ts">
defineProps<{
  currencyCode: CurrencyCode,
}>()
</script>

<template>
  <BcIconFlagAu v-if="currencyCode === 'AUD'" />
  <BcIconFlagCa v-if="currencyCode === 'CAD'" />
  <BcIconFlagCn v-if="currencyCode === 'CNY'" />
  <BcIconFlagEu v-if="currencyCode === 'EUR'" />
  <BcIconFlagGb v-if="currencyCode === 'GBP'" />
  <BcIconFlagJp v-if="currencyCode === 'JPY'" />
  <BcIconFlagUs v-if="currencyCode === 'USD'" />
  <BcIconCrypto
    v-if="isCrypto(currencyCode)"
    :currency-code
    color-mode="currentColor"
  />
</template>
