<script setup lang="ts">
import type { ModelRef } from 'vue'
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import { ChainInfo, ChainID } from '~/types/network'

const { availableNetworks, isNetworkDisabled } = useNetworkStore()
const { bridgeArraysRefs, bridgePrimitiveRefs } = useRefBridge()

/** This ref is a chain ID if only one network can be selected, or an array of chain IDs if several networks can be selected. */
const liveState = defineModel<ChainID|ChainID[]>()

let selectionMulti : ModelRef<string[]>
const selectionSingle = ref<string>()

const buttons = computed(() => {
  const list: MultiBarItem[] = []
  availableNetworks.value.forEach(chainId => list.push({
    component: IconNetwork,
    componentProps: { chainId, harmonizePerceivedSize: true, colored: true },
    componentClass: 'maximum',
    value: String(chainId),
    disabled: isNetworkDisabled(chainId),
    tooltip: ChainInfo[chainId].name + ' ' + ChainInfo[chainId].description
  }))
  return list
})

if (Array.isArray(liveState.value)) {
  bridgeArraysRefs(liveState as ModelRef<ChainID[]>, selectionMulti)
} else {
  bridgePrimitiveRefs(liveState as ModelRef<ChainID>, selectionSingle)
}
</script>

<template>
  <BcToggleMultiBar v-if="selectionMulti" v-model="selectionMulti" :buttons="buttons" />
  <BcToggleSingleBar v-if="selectionSingle" v-model="selectionSingle" :buttons="buttons" />
</template>

<style lang="scss">
.maximum {
  width: 100%;
  height: 100%;
}
</style>
