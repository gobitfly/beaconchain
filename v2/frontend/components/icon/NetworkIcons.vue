<script setup lang="ts">
import { ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({
  chainId: { type: Number, required: true }
})

const family = ChainInfo[props.chainId as ChainIDs].family
</script>

<template>
  <IconNetworkAnyIcon v-if="family === 'Any'" />
  <IconNetworkEthereumIcon v-else-if="family === 'Ethereum'" :background-theme="useColorMode().value" />
  <IconNetworkArbitrumIcon v-else-if="family === 'Arbitrum'" />
  <IconNetworkOptimismIcon v-else-if="family === 'Optimism'" />
  <IconNetworkBaseIcon v-else-if="family === 'Base'" />
  <IconNetworkGnosisIcon v-else-if="family === 'Gnosis'" />
</template>
