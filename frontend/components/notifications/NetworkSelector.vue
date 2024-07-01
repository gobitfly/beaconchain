<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import { ChainInfo } from '~/types/network'

const { availableNetworks } = useNetworkStore()

const selection = ref<string[]>(['1'])

const buttons = computed(() => {
  const list: MultiBarItem[] = []
  availableNetworks.value.forEach(chainId => list.push({
    component: IconNetwork,
    componentProps: { chainId, harmonizePerceivedSize: true, colored: true },
    value: String(chainId),
    tooltip: ChainInfo[chainId].name + ' ' + ChainInfo[chainId].description
  }))
  return list
})
</script>

<template>
  <BcToggleMultiBar v-model="selection" :icons="buttons" />
</template>

<style scoped lang="scss">

</style>
