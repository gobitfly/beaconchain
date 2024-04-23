<script setup lang="ts">
import { SearchbarStyle, type NetworkFilter } from '~/types/searchbar'
import { ChainInfo, ChainIDs } from '~/types/networks'

const emit = defineEmits<{(e: 'change') : void}>()
defineProps<{
  barStyle : SearchbarStyle
}>()
const liveState = defineModel<NetworkFilter>({ required: true }) // each entry has a Category as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

const { t } = useI18n()

const head = ref<{look : 'on'|'off', network : string}>({
  look: 'off',
  network: ''
})
const listInDropdown = ref<{
  chainId: ChainIDs,
  label: string,
  selected: boolean
}[]>([])
const dropdownIsOpen = ref<boolean>(false)

watch(liveState, () => updateLocalState()) // fires when the parent changes the pointer but not when he or we change a value inside

onBeforeMount(() => {
  dropdownIsOpen.value = false
  updateLocalState()
})

function updateLocalState () {
  // first we update the head
  let howManyAreSelected = 0
  for (const nw of liveState.value) {
    if (nw[1]) {
      howManyAreSelected++
    }
  }
  const allNetworksAreSelected = (howManyAreSelected === liveState.value.size)
  if (howManyAreSelected === 0 || allNetworksAreSelected) {
    head.value.network = t('search_bar.all_networks')
  } else {
    head.value.network = String(howManyAreSelected)
  }
  head.value.look = (howManyAreSelected === 0) ? 'off' : 'on'
  // now the listInDropdown
  listInDropdown.value.length = 0
  listInDropdown.value.push({ chainId: ChainIDs.Any, label: t('search_bar.all_networks'), selected: allNetworksAreSelected })
  for (const filter of liveState.value) {
    listInDropdown.value.push({ chainId: filter[0], label: ChainInfo[filter[0]].description, selected: filter[1] })
  }
}

function oneOptionChanged (index : number) {
  const selected = listInDropdown.value[index].selected
  if (listInDropdown.value[index].chainId !== ChainIDs.Any) {
    liveState.value.set(listInDropdown.value[index].chainId, selected)
  } else {
    for (const filter of liveState.value) {
      liveState.value.set(filter[0], selected)
    }
  }
  updateLocalState()
  emit('change')
}
</script>

<template>
  <div class="anchor">
    <BcSearchbarFilterButton
      class="head"
      :bar-style="barStyle"
      :look="head.look"
      :state="dropdownIsOpen"
      @change="(open : boolean) => dropdownIsOpen = open"
    >
      <div class="content">
        <span class="label">
          {{ t('search_bar.network_filter_label') + ' ' + head.network }}
        </span>
        <span class="arrow">
          ▾
        </span>
      </div>
    </BcSearchbarFilterButton>
    <div
      v-if="dropdownIsOpen"
      class="dropdown"
      @click="(e : Event) => e.stopPropagation()"
      @mouseleave="dropdownIsOpen = false"
    >
      <div v-for="(item, i) of listInDropdown" :key="item.chainId" class="line" @click="oneOptionChanged(i)">
        <Checkbox v-model="item.selected" :binary="true" :input-id="String(item.chainId)" />
        <label :for="String(item.chainId)">
          {{ item.label }}
        </label>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.anchor {
  display: inline-block;
    padding-bottom: 8px;

  .head {
    position: relative;
    width: 94px;
    .content {
      position: relative;
      display: inline-flex;
      width: 100%;
      .label {
        display: inline-flex;
        flex-grow: 1;
      }
      .arrow {
        display: inline-flex;
      }
    }
  }

  .dropdown {
    position: absolute;
    z-index: 1024;
    width: 128px;
    border-radius: 10px;
    left: 0px;
    top: 21px;
    background-color: var(--light-grey);
    @include fonts.small_text_bold;
    color: var(--light-black);

    .line {
      position:relative;
      cursor: pointer;
    }
  }
}
</style>
