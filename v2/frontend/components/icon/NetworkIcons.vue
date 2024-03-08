<script setup lang="ts">
import { ChainFamily, ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({
  chainId: { type: Number, required: true },
  colored: { type: Boolean, default: false },
  harmonizePerceivedSize: { type: Boolean, default: false }
})

const family = ChainInfo[props.chainId as ChainIDs].family
const harmonizationClass = props.harmonizePerceivedSize ? family : ''
</script>

<template>
  <span>
    <span v-if="family === ChainFamily.Any" />
    <span v-else-if="props.colored" class="container" :class="harmonizationClass">
      <IconNetworkEthereumMulti v-if="family === ChainFamily.Ethereum" />
      <IconNetworkArbitrumMulti v-else-if="family === ChainFamily.Arbitrum" />
      <IconNetworkOptimismMulti v-else-if="family === ChainFamily.Optimism" />
      <IconNetworkBaseMulti v-else-if="family === ChainFamily.Base" />
      <IconNetworkGnosisMulti v-else-if="family === ChainFamily.Gnosis" />
    </span>
    <span v-else class="container" :class="harmonizationClass">
      <IconNetworkEthereumMono v-if="family === ChainFamily.Ethereum" />
      <IconNetworkArbitrumMono v-else-if="family === ChainFamily.Arbitrum" />
      <IconNetworkOptimismMono v-else-if="family === ChainFamily.Optimism" />
      <IconNetworkBaseMono v-else-if="family === ChainFamily.Base" />
      <IconNetworkGnosisMono v-else-if="family === ChainFamily.Gnosis" />
    </span>
  </span>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border: none;
  max-height: 100%;
}

// the following classes are used only if props `harmonize-perceived-size` has been set to true

.Ethereum { // due to its height, the logo looks bigger than the others
  width: 90%;
  margin: auto;
}

.Arbitrum {  // due to the round border in its design, the logo looks smaller than the others
  position: relative;
  width: 110%;
  left: -1px; // this shift is important when the icon is small and without visible consequence when it is big
}
</style>
