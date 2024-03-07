<script setup lang="ts">
import { ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({
  chainId: { type: Number, required: true },
  coloured: { type: Boolean, default: false }
})

const family = ChainInfo[props.chainId as ChainIDs].family
</script>

<template>
  <span v-if="family === 'Any'" />
  <span v-else-if="props.coloured">
    <IconNetworkEthereumMulti v-if="family === 'Ethereum'" class="fill-the-container" />
    <IconNetworkArbitrumMulti v-else-if="family === 'Arbitrum'" class="fill-the-container" />
    <IconNetworkOptimismMulti v-else-if="family === 'Optimism'" class="fill-the-container" />
    <IconNetworkBaseMulti v-else-if="family === 'Base'" class="fill-the-container" />
    <IconNetworkGnosisMulti v-else-if="family === 'Gnosis'" class="fill-the-container" />
  </span>
  <span v-else>
    <IconNetworkEthereumMono v-if="family === 'Ethereum'" class="fill-the-container" />
    <IconNetworkArbitrumMono v-else-if="family === 'Arbitrum'" class="fill-the-container" />
    <IconNetworkOptimismMono v-else-if="family === 'Optimism'" class="fill-the-container" />
    <IconNetworkBaseMono v-else-if="family === 'Base'" class="fill-the-container" />
    <IconNetworkGnosisMono v-else-if="family === 'Gnosis'" class="fill-the-container" />
  </span>
</template>

<style lang="scss" scoped>
.fill-the-container {
  width: 100%;
  height: 100%;
}
</style>
