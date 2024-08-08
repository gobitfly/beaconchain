<script setup lang="ts">
import type { CryptoCurrency } from '~/types/currencies'

interface Props {
  currency?: CryptoCurrency
}
defineProps<Props>()
</script>

<template>
  <IconNetworkEthereum
    v-if="currency === 'ETH'"
    class="monochromatic"
  />
  <IconNetworkGnosis
    v-else-if="currency === 'GNO'"
    class="monochromatic"
  />
  <IconCryptoDai
    v-else-if="currency === 'DAI' || currency === 'xDAI'"
    class="monochromatic"
  />
</template>
