<script setup lang="ts">
import { SearchbarStyle } from '~/types/searchbar'
import { ChainIDs, ChainInfo } from '~/types/networks'

const emit = defineEmits(['change'])
const props = defineProps<{
  initialState: Record<string, boolean>, // each key is a stringifyed chain ID (as enumerated in ChainIDs in networks.ts)
  barStyle: SearchbarStyle
}>()

let vueMultiselectAllOptions : {name: string, label: string}[] = []
const vueMultiselectSelectedOptions = ref<string[]>([])

let componentIsReady = false
const state : Record<string, boolean> = {} // each key is a stringifyed chain ID (as enumerated in ChainIDs in networks.ts)
const everyNetworkIsSelected = ref<boolean>(false)

onMounted(() => {
  componentIsReady = false

  vueMultiselectAllOptions = []
  vueMultiselectSelectedOptions.value = []

  for (const nw in props.initialState) {
    state[nw] = props.initialState[nw]
    vueMultiselectAllOptions.push({ name: nw, label: ChainInfo[Number(nw) as ChainIDs].description })
    if (state[nw]) {
      vueMultiselectSelectedOptions.value.push(nw)
    }
  }
  everyNetworkIsSelected.value = (vueMultiselectSelectedOptions.value.length === vueMultiselectAllOptions.length)

  componentIsReady = true
})

function selectionHasChanged () {
  if (!componentIsReady) {
    // ensures that we do not emit change-events during the initialization of the drop-down (see the code in onMounted)
    return
  }
  console.log('Network selector')
  everyNetworkIsSelected.value = (vueMultiselectSelectedOptions.value.length === vueMultiselectAllOptions.length)
  for (const nw in state) {
    state[nw] = vueMultiselectSelectedOptions.value.includes(nw)
  }
  emit('change', state)
}
</script>

<template>
  <!--do not remove '&nbsp;' in the placeholder otherwise the CSS of the component believes that nothing is selected when everthing is selected-->
  <MultiSelect
    v-model="vueMultiselectSelectedOptions"
    :options="vueMultiselectAllOptions"
    option-value="name"
    option-label="label"
    placeholder="Networks:&nbsp;all"
    :variant="'filled'"
    display="comma"
    :show-toggle-all="false"
    :max-selected-labels="1"
    :selected-items-label="'Networks: ' + (everyNetworkIsSelected ? 'all' : '{0}')"
    append-to="self"
    @change="selectionHasChanged"
    @click="(e : Event) => e.stopPropagation()"
  />
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/fonts.scss";

.p-multiselect {
  @include fonts.small_text_bold;
  width: 128px;
  height: 20px;
  border-radius: 10px;

  .p-multiselect-trigger {
    width: 1.5rem;
  }
  .p-multiselect-label {
    padding-top: 3px;
    border-top-left-radius: 10px;
    border-bottom-left-radius: 10px;
    .p-placeholder {
      border-top-left-radius: 10px;
      border-bottom-left-radius: 10px;
      background: var(--searchbar-filter-unselected-gaudy);
    }
  }
  &.p-multiselect-panel {
    width: 140px;
    max-height: 100px;
    overflow: auto;
  }
}
</style>
