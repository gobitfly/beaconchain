<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
import { IconNetwork } from '#components'
import { ChainInfo, ChainID } from '~/types/network'

const props = defineProps<{
  displayOnly?: boolean
}>()

const { availableNetworks, isNetworkDisabled } = useNetworkStore()

/** If the v-model is:
 *  - A ChainID: only one network can be selected by the user. Prop `:display-only` must be false or omitted.
 *  - An array of ChainID:
 *    - and prop `:display-only` is `false`/omitted: several networks can be selected by the user,
 *    - and prop `:display-only` is `true`: the networks in the array are shown to the user but they are unclickable. */
const liveState = defineModel<ChainID|ChainID[]>({ required: false })

const selection = Array.isArray(liveState.value)
  ? useArrayRefBridge<ChainID, string>(liveState as Ref<ChainID[]>, true)
  : usePrimitiveRefBridge<ChainID, string>(liveState as Ref<ChainID>)

const buttons = shallowRef<MultiBarItem[]>([])

if (props.displayOnly) {
  watch(liveState as Ref<ChainID[]>, updateButtons, { immediate: true })
} else {
  watch(availableNetworks, updateButtons, { immediate: true })
}

function updateButtons (source: ChainID[]) : void {
  buttons.value = []
  source.forEach((chainId) => {
      buttons.value!.push({
        component: IconNetwork,
        componentProps: { chainId, harmonizePerceivedSize: true, colored: true },
        componentClass: 'maximum',
        value: String(chainId),
        disabled: isNetworkDisabled(chainId),
        tooltip: ChainInfo[chainId].name + ' ' + ChainInfo[chainId].description
      })
  })
}
</script>

<template>
  <BcToggleMultiBar v-if="Array.isArray(selection)" v-model="selection" :buttons="buttons" :display-mode="displayOnly" />
  <BcToggleSingleBar v-else v-model="selection" :buttons="buttons" layout="minimal" />
</template>

<style lang="scss">
.maximum {
  width: 100%;
  height: 100%;
}
</style>
