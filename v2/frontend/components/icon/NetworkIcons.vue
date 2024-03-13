<script setup lang="ts">
import { ChainFamily, ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({
  chainId: { type: Number, required: true },
  colored: { type: Boolean, default: false },
  harmonizePerceivedSize: { type: Boolean, default: false }
})

const family = ChainInfo[props.chainId as ChainIDs].family
const sizing = props.harmonizePerceivedSize ? family : 'do-not-adjust-size'
const coloring = !props.colored ? 'monochromatic' : 'do-not-change-colors'
</script>

<template>
  <span>
    <span class="container" :class="sizing">
      <IconNetworkEthereumColored v-if="family === ChainFamily.Ethereum" :class="coloring" />
      <IconNetworkArbitrumColored v-else-if="family === ChainFamily.Arbitrum" :class="[coloring, sizing]" />
      <IconNetworkOptimismColored v-else-if="family === ChainFamily.Optimism" :class="[coloring, sizing]" />
      <IconNetworkBaseColored v-else-if="family === ChainFamily.Base" :class="[coloring, sizing]" />
      <IconNetworkGnosisColored v-else-if="family === ChainFamily.Gnosis" :class="[coloring, sizing]" />
    </span>
  </span>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  max-height: 100%;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border: none;
}
// The following classes are used only if props `harmonize-perceived-size` has been set to true.
// It correct what a human brain perceives, to give the feeling that all icons have a similar size (maybe a matter of surface area).
// Based on empirical trials an errors with author's brain.

.Ethereum {
  margin: auto;
  width: 85%;
}

.Arbitrum {
  position: relative;
  width: 105%;
}
</style>
