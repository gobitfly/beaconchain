<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import { ChainInfo, ChainID } from '~/types/network'

const { availableNetworks, isNetworkDisabled } = useNetworkStore()

/**
 * If `v-model:singleselect="..."` is in the props, then the user can select only one network.
 * If `v-model:multiselect="..."` is in the props, then the user can select several networks.
 */
const liveStateMulti = defineModel<ChainID[]>('multiselect')
const liveStateSingle = defineModel<ChainID>('singleselect')

const selectionMulti = ref<string[]>([])
const selectionSingle = ref<string>('')

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

watch(liveStateMulti, (input: ChainID[]|undefined) => {
  if (!input) {
    selectionMulti.value = []
    return
  }
  if (JSON.stringify(input) !== JSON.stringify(selectionMulti.value)) {
    selectionMulti.value = input.map(id => String(id))
  }
})

watch(liveStateSingle, (input: ChainID|undefined) => {
  if (!input) {
    selectionSingle.value = ''
    return
  }
  if (String(input) !== selectionSingle.value) {
    selectionSingle.value = String(input)
  }
})

watch(selectionMulti, (output: string[]) => {
  if (JSON.stringify(output) !== JSON.stringify(liveStateMulti.value)) {
    liveStateMulti.value = output.map(id => Number(id) as ChainID)
  }
})

watch(selectionSingle, (output: string) => {
  if (Number(output) as ChainID !== liveStateSingle.value) {
    liveStateSingle.value = Number(output) as ChainID
  }
})
</script>

<template>
  <BcToggleMultiBar v-if="liveStateMulti" v-model="selectionMulti" :buttons="buttons" />
  <BcToggleSingleBar v-if="liveStateSingle" v-model="selectionSingle" :buttons="buttons" />
</template>

<style lang="scss">
.maximum {
  width: 100%;
  height: 100%;
}
</style>
