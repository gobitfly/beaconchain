<script setup lang="ts">
defineProps<{
  currencyCode: CurrencyCode,
}>()
</script>

<template>
  <LazyBcIconFlagAu v-if="currencyCode === 'AUD'" />
  <LazyBcIconFlagCa v-if="currencyCode === 'CAD'" />
  <LazyBcIconFlagCn v-if="currencyCode === 'CNY'" />
  <LazyBcIconFlagEu v-if="currencyCode === 'EUR'" />
  <LazyBcIconFlagGb v-if="currencyCode === 'GBP'" />
  <LazyBcIconFlagJp v-if="currencyCode === 'JPY'" />
  <LazyBcIconFlagUs v-if="currencyCode === 'USD'" />
  <LazyBcIconCrypto
    v-if="isCrypto(currencyCode)"
    :currency-code
    color-mode="currentColor"
  />
</template>
