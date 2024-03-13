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
  <IconNetworkEthereumColored v-if="family === ChainFamily.Ethereum" :class="[coloring, sizing]" />
  <IconNetworkArbitrumColored v-else-if="family === ChainFamily.Arbitrum" :class="[coloring, sizing]" />
  <IconNetworkOptimismColored v-else-if="family === ChainFamily.Optimism" :class="[coloring, sizing]" />
  <IconNetworkBaseColored v-else-if="family === ChainFamily.Base" :class="[coloring, sizing]" />
  <IconNetworkGnosisColored v-else-if="family === ChainFamily.Gnosis" :class="[coloring, sizing]" />
</template>

<style lang="scss" scoped>
// the following classes are used only if props `harmonize-perceived-size` has been set to true

.Ethereum { // due to its height, the logo looks bigger than the others
  position: relative;
  display: flex;
  margin: auto;
  height: 90%;
}

.Arbitrum {  // due to the round border in its design, the logo looks smaller than the others
  position: relative;
  width: 110%;
}
</style>
