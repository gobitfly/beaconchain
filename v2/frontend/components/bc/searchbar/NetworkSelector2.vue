<script setup lang="ts">
import { SearchbarStyle, type NetworkFilter } from '~/types/searchbar'
// import { ChainIDs, ChainInfo } from '~/types/networks'

// const emit = defineEmits<{(e: 'change') : void}>()
defineProps<{
  barStyle : SearchbarStyle
}>()
const liveState = defineModel<NetworkFilter>({ required: true }) // each entry has a Category as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

const { t } = useI18n()

const dropDownIsOpen = ref<boolean>(false)

onUnmounted(() => {
  dropDownIsOpen.value = false
})

const everyNetworkIsSelected = computed(() => {
  for (const nw of liveState.value) {
    if (!nw[1]) {
      return false
    }
  }
  return true
})

/*
function selectionHasChanged (chainId : ChainIDs, selected : boolean) {
  liveState.value.set(chainId, selected) // the map element cannot save the infornmation from `v-model`, so we do it with .set()
  emit('change')
}
*/
</script>

<template>
  <div>
    <BcSearchbarMiniButton
      v-model="dropDownIsOpen"
      class="button"
      :bar-style="barStyle"
      :forced-color="1"
    >
      {{ t('search_bar.network_filter_label') + ' ' + (everyNetworkIsSelected ? t('search_bar.all_networks') : '{0}') }}
    </BcSearchbarMiniButton>
  </div>
</template>

<style lang="scss" scoped>
.button {
  margin-bottom: 8px;
}
</style>
