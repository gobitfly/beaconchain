<script setup lang="ts">
interface Props {
  currency?: CurrencyCodeCrypto,
}
defineProps<Props>()
</script>

<template>
  <IconNetworkEthereum
    v-if="currency === 'ETH'"
    class="monochromatic"
  />

  <IconCryptoDai
    v-else-if="currency === 'DAI' || currency === 'xDAI'"
    class="monochromatic"
  />
</template>
