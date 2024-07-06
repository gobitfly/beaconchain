<script setup lang="ts">
import { ChainFamily, ChainIDs, ChainInfo } from '~/types/network'
const colorMode = useColorMode()

// Usage :
// Both properties width and height must be set by the parent component. The icon will fill this frame as much as possible without deformation.
// Additional properties are the following:
const props = defineProps({
  chainId: { type: Number, required: true }, // network whose icon must be displayed (L2s and tesnets can be given)
  colored: { type: Boolean, default: false }, // tells whether the icon must be in its official color or in the color of the font
  doNotAdaptToColorTheme: { type: Boolean, default: false }, // some icons in their original colors are hard to see on a dark background, so they are adapted automatically, but you can deactivate this behavior if your background is always light
  harmonizePerceivedSize: { type: Boolean, default: false } // makes some icons slightly smaller/bigger, to appear with a size "similar" to the others
})

const family = computed(() => ChainInfo[props.chainId as ChainIDs].family)
const coloring = computed(() => !props.colored ? 'monochromatic' : (colorMode.value !== 'dark' || props.doNotAdaptToColorTheme ? '' : 'pastel'))
const sizing = computed(() => props.harmonizePerceivedSize ? family.value : '')
</script>

<template>
  <div class="frame">
    <IconNetworkEthereum v-if="family === ChainFamily.Ethereum" class="icon" :class="[sizing,coloring]" />
    <IconNetworkArbitrum v-else-if="family === ChainFamily.Arbitrum" class="icon" :class="[sizing,coloring]" />
    <IconNetworkOptimism v-else-if="family === ChainFamily.Optimism" class="icon" :class="[sizing,coloring]" />
    <IconNetworkBase v-else-if="family === ChainFamily.Base" class="icon" :class="[sizing,coloring]" />
    <IconNetworkGnosis v-else-if="family === ChainFamily.Gnosis" class="icon" :class="[sizing,coloring]" />
  </div>
</template>

<style lang="scss" scoped>
.frame {
  position: relative;
  display: inline-block;
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
  // Based on empirical trials an errors with author's brain.
  .Ethereum {
    height: 110%;
  }

  .Arbitrum {
    height: 90%;
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
}
</style>
