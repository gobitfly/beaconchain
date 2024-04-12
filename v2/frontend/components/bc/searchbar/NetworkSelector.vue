<script setup lang="ts">
import { SearchbarStyle } from '~/types/searchbar'
import { ChainIDs, ChainInfo } from '~/types/networks'

const emit = defineEmits<{(e: 'change') : void}>()
const props = defineProps<{
  liveState: Record<string, boolean>, // each key is a stringifyed chain ID (as enumerated in ChainIDs in networks.ts) and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.
  barStyle: SearchbarStyle
}>()

const { t } = useI18n()

const vueMultiselectAllOptions = ref<{name: string, label: string}[]>()
const vueMultiselectSelectedOptions = ref<string[]>([])

let statePtr : Record<string, boolean> = {} // each key is a stringifyed chain ID and the state of the option as value

const everyNetworkIsSelected = computed(() => {
  return (vueMultiselectSelectedOptions.value.length === vueMultiselectAllOptions.value?.length)
})

initialize()

watch(props, () => {
  initialize()
})

function initialize () {
  vueMultiselectAllOptions.value = []
  vueMultiselectSelectedOptions.value = []
  statePtr = props.liveState

  for (const nw in statePtr) {
    vueMultiselectAllOptions.value.push({ name: nw, label: ChainInfo[Number(nw) as ChainIDs].description })
    if (statePtr[nw]) {
      vueMultiselectSelectedOptions.value.push(nw)
    }
  }
}

function selectionHasChanged () {
  for (const nw in statePtr) {
    statePtr[nw] = vueMultiselectSelectedOptions.value.includes(nw)
  }
  emit('change')
}
</script>

<template>
  <!--do not remove '&nbsp;' in the placeholder otherwise the CSS of the component believes that nothing is selected when everthing is selected-->
  <MultiSelect
    v-model="vueMultiselectSelectedOptions"
    :options="vueMultiselectAllOptions"
    option-value="name"
    option-label="label"
    :placeholder="t('search_bar.network_filter_label')+'&nbsp;'+t('search_bar.all_networks')"
    :variant="'filled'"
    display="comma"
    :show-toggle-all="false"
    :max-selected-labels="1"
    :selected-items-label="t('search_bar.network_filter_label') + ' ' + (everyNetworkIsSelected ? t('search_bar.all_networks') : '{0}')"
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
  margin-bottom: 8px;

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
