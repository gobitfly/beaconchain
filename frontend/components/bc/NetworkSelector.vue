<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import {
  type ChainIDs, ChainInfo,
} from '~/types/network'

const props = defineProps<{
  readonlyNetworks?: ChainIDs[],
}>()

const {
  availableNetworks, isNetworkDisabled,
} = useNetworkStore()

/** If prop `:readonly-networks` is given:
 *   the networks in array `:readonly-networks` are shown to the user and they are unclickable,
 *  Otherwise, give a v-model. If the v-model is
 *   a ChainIDs: only one network can be selected by the user,
 *   an array of ChainIDs: several networks can be selected by the user */
const selection = defineModel<ChainIDs | ChainIDs[]>({ required: false })

let barSelection: Ref<string> | Ref<string[]>
if (props.readonlyNetworks) {
  barSelection = ref<string[]>([])
}
else if (Array.isArray(selection.value)) {
  barSelection = useArrayRefBridge<ChainIDs, string>(
    selection as Ref<ChainIDs[]>,
  )
}
else {
  barSelection = usePrimitiveRefBridge<ChainIDs, string>(
    selection as Ref<ChainIDs>,
  )
}

const buttons = computed(() => {
  const list: MultiBarItem[] = []
  const source = props.readonlyNetworks || availableNetworks.value
  for (const chainId of source) {
    list.push({
      component: IconNetwork,
      componentClass: 'maximum',
      componentProps: {
        chainId,
        colored: true,
        harmonizePerceivedSize: true,
      },
      disabled: isNetworkDisabled(chainId),
      tooltip: ChainInfo[chainId].name,
      value: String(chainId),
    })
  }
  return list
})
</script>

<template>
  <BcToggleMultiBar
    v-if="Array.isArray(barSelection)"
    v-model="barSelection"
    :buttons
    :readonly-mode="!!readonlyNetworks"
  />
  <BcToggleSingleBar
    v-else
    v-model="barSelection"
    :buttons
    layout="minimal"
  />
</template>

<style lang="scss">
.maximum {
  width: 100%;
  height: 100%;
}
</style>
