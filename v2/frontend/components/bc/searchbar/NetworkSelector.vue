<script setup lang="ts">
import { SearchbarStyle, type NetworkFilter } from '~/types/searchbar'
import { ChainInfo, ChainIDs } from '~/types/networks'

const emit = defineEmits<{(e: 'change') : void}>()
defineProps<{
  barStyle: SearchbarStyle
}>()
const liveState = defineModel<NetworkFilter>({ required: true }) // each entry has a ChainIDs as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

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

watch(liveState, updateLocalState) // fires when the parent changes the whole object but not when he / we change a value inside

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
  // now we update the list used to fill the dropdown
  listInDropdown.value.length = 0
  listInDropdown.value.push({ chainId: ChainIDs.Any, label: t('search_bar.all_networks'), selected: allNetworksAreSelected })
  for (const filter of liveState.value) {
    listInDropdown.value.push({ chainId: filter[0], label: ChainInfo[filter[0]].name, selected: filter[1] })
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
        ▾
      </div>
    </BcSearchbarFilterButton>
    <div
      v-if="dropdownIsOpen"
      class="dropdown"
      @click="(e : Event) => e.stopPropagation()"
    >
      <div v-for="(line, i) of listInDropdown" :key="line.chainId" class="line" @click="oneOptionChanged(i)">
        <Checkbox v-model="line.selected" :binary="true" :input-id="String(line.chainId)" />
        <label :for="String(line.chainId)" class="label">
          {{ line.label }}
        </label>
        <IconNetwork :chain-id="line.chainId" :colored="true" :harmonize-perceived-size="true" class="icon" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.anchor {
  display: inline-block;
  padding-bottom: 8px;
  box-sizing: border-box;

  .head {
    position: relative;
    .content {
      position: relative;
      display: inline-flex;
      width: 85px;
      .label {
        display: inline-flex;
        flex-grow: 1;
      }
    }
  }

  .dropdown {
    position: absolute;
    display: block;
    box-sizing: border-box;
    z-index: 1024;
    border-radius: 10px;
    left: 0px;
    top: 21px;
    padding: var(--padding);
    background-color: var(--light-grey);
    @include fonts.small_text_bold;
    color: var(--light-black);

    .line {
      position:relative;
      display: flex;
      width: 100%;
      margin-bottom: 2px;
      white-space: nowrap;

      .p-checkbox {
        :deep(.p-checkbox-box:not(:hover):not(.p-highlight)) {
          background: var(--light-grey-3);
        }
      }

      .label {
        position:relative;
        display: inline-flex;
        flex-grow: 1;
        margin-left: 5px;
        margin-top: auto;
        margin-bottom: auto;
        cursor: pointer;
        user-select: none;
      }

      .icon {
        box-sizing: border-box;
        position: relative;
        margin-left: 10px;
        width: 18px;
        height: 18px;
        right: 0px;
        top: 1px;
      }
    }
  }
}
</style>
