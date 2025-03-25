<script setup lang="ts">
const { currencyCode } = defineProps<{
  currencyCode: CurrencyCode,
}>()
</script>

<template>
  <BcIcon
    v-if="currencyCode === 'AUD'"
    name="flag-au"
  />
  <BcIcon
    v-else-if="currencyCode === 'CAD'"
    name="flag-ca"
  />
  <BcIcon
    v-else-if="currencyCode === 'CNY'"
    name="flag-cn"
  />
  <BcIcon
    v-else-if="currencyCode === 'EUR'"
    name="flag-eu"
  />
  <BcIcon
    v-else-if="currencyCode === 'GBP'"
    name="flag-gb"
  />
  <BcIcon
    v-else-if="currencyCode === 'JPY'"
    name="flag-jp"
  />
  <BcIcon
    v-else-if="currencyCode === 'USD'"
    name="flag-us"
  />

  <BcIcon
    v-else-if="currencyCode === 'xDAI' || currencyCode === 'DAI'"
    name="crypto-xdai"
  />
  <BcIcon
    v-else-if="currencyCode === 'GNO' || currencyCode === 'mGNO'"
    name="crypto-gno"
  />
  <BcIcon
    v-else-if="currencyCode === 'ETH'"
    name="crypto-eth"
  />
</template>
