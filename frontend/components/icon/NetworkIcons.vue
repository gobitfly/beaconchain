<script setup lang="ts">
import { ChainFamily, ChainIDs, ChainInfo } from '~/types/networks'

// Usage :
// Both properties width and height must be set by the parent component. The icon will fill this frame as much as possible without deformation.
// Additional properties are the following:
const props = defineProps({
  chainId: { type: Number, required: true }, // network whose icon must be displayed (L2s and tesnets can be given)
  colored: { type: Boolean, default: false }, // tells whether the icon must be in its official color or in the color of the font
  harmonizePerceivedSize: { type: Boolean, default: false } // makes certain icons slightly smaller, not to appear bigger than the others
})

const family = ChainInfo[props.chainId as ChainIDs].family
const sizing = props.harmonizePerceivedSize ? family : 'do-not-adjust-size'
const coloring = !props.colored ? 'monochromatic' : 'do-not-change-colors'
</script>

<template>
  <div class="frame">
    <IconNetworkEthereumColored v-if="family === ChainFamily.Ethereum" class="icon" :class="[sizing,coloring]" />
    <IconNetworkArbitrumColored v-else-if="family === ChainFamily.Arbitrum" class="icon" :class="[sizing,coloring]" />
    <IconNetworkOptimismColored v-else-if="family === ChainFamily.Optimism" class="icon" :class="[sizing,coloring]" />
    <IconNetworkBaseColored v-else-if="family === ChainFamily.Base" class="icon" :class="[sizing,coloring]" />
    <IconNetworkGnosisColored v-else-if="family === ChainFamily.Gnosis" class="icon" :class="[sizing,coloring]" />
  </div>
</template>

<style lang="scss" scoped>
.frame {
  position: relative;
  display: inline-block;
}

.icon {
  position: absolute;
  height: 100%;
  width: 100%;
  top: 50%;
  left: 50%;
  transform: translate(-50%,-50%);
}

// The following classes are used only if props `harmonize-perceived-size` has been set to true.
// It corrects what a human brain perceives, to give the feeling that all icons have a similar size (maybe a matter of surface area).
// Based on empirical trials an errors with author's brain, works better on a dark background.

.Ethereum {
  height: 100%;
}

.Arbitrum {
  height: 95%;
}

.Optimism {
  height: 90%;
}

.Base {
  height: 90%;
}

.Gnosis {
  height: 90%;
}
</style>
