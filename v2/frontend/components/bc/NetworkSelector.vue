<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import { ChainInfo, ChainID } from '~/types/network'

const { availableNetworks, isNetworkDisabled } = useNetworkStore()

/** This ref is a chain ID if only one network can be selected, or an array of chain IDs if several networks can be selected. */
const liveState = defineModel<ChainID|ChainID[]>()

const selection = Array.isArray(liveState.value)
  ? useArrayRefBridge<ChainID, string>(liveState as Ref<ChainID[]>)
  : usePrimitiveRefBridge<ChainID, string>(liveState as Ref<ChainID>)

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
</script>

<template>
  <BcToggleMultiBar v-if="Array.isArray(selection)" v-model="selection" :buttons="buttons" />
  <BcToggleSingleBar v-else v-model="selection" :buttons="buttons" />
</template>

<style lang="scss">
.maximum {
  width: 100%;
  height: 100%;
}
</style>
