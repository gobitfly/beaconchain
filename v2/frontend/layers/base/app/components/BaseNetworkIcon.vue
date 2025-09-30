<script setup lang="ts">
const { chainId } = defineProps<{
  chainId: ChainId,
}>()

const shortName = computed(() => getNetworkShortName(chainId))
</script>

<template>
  <LazyNetworkIconCryptoEth
    v-if="shortName === 'Ethereum'"
    class="w-2xl"
  />
  <LazyNetworkIconCryptoHoodi
    v-if="shortName === 'Hoodi'"
    class="w-2xl"
  />
</template>
