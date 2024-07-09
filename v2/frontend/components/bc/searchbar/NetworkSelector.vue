<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCaretDown } from '@fortawesome/pro-solid-svg-icons'
import { SearchbarShape, type SearchbarColors, type SearchbarDropdownLayout, type NetworkFilter } from '~/types/searchbar'
import { ChainInfo, ChainIDs } from '~/types/network'

const emit = defineEmits<{(e: 'change') : void}>()
defineProps<{
  barShape: SearchbarShape,
  colorTheme: SearchbarColors,
  dropdownLayout : SearchbarDropdownLayout
}>()
const liveState = defineModel<NetworkFilter>({ required: true }) // each entry has a ChainIDs as key and the state of the option as value. The component will write directly into it, so the data of the parent is always up-to-date.

const { t } = useI18n()

const headState = ref<{look : 'on'|'off', network : string}>({
  look: 'off',
  network: ''
})
const listInDropdown = ref<{
  chainId: ChainIDs,
  label: string,
  selected: boolean
}[]>([])
const dropdownIsOpen = ref<boolean>(false)

const head = ref<HTMLDivElement>()
const dropdown = ref<HTMLDivElement>()

watch(liveState, updateLocalState) // fires when the parent changes the whole object but not when he / we change a value inside

onBeforeMount(() => {
  dropdownIsOpen.value = false
  updateLocalState()
  document.addEventListener('keydown', listenToKeys)
  document.addEventListener('click', listenToClicks)
})

onUnmounted(() => {
  document.removeEventListener('click', listenToClicks)
  document.removeEventListener('keydown', listenToKeys)
})

function listenToClicks (event : Event) {
  if (!dropdownIsOpen.value || !dropdown.value || !head.value || dropdown.value.contains(event.target as Node) || head.value.contains(event.target as Node)) {
    return
  }
  dropdownIsOpen.value = false
}

function listenToKeys (event : KeyboardEvent) {
  if (event.key === 'Escape' && dropdownIsOpen.value) {
    dropdownIsOpen.value = false
    event.stopImmediatePropagation()
  }
}

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
    headState.value.network = t('search_bar.all_networks')
  } else {
    headState.value.network = String(howManyAreSelected)
  }
  headState.value.look = (howManyAreSelected === 0) ? 'off' : 'on'
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
      :bar-shape="barShape"
      :color-theme="colorTheme"
      :dropdown-layout="dropdownLayout"
      :look="headState.look"
      :state="dropdownIsOpen"
      @change="(open : boolean) => dropdownIsOpen = open"
    >
      <div ref="head" class="content">
        <span class="label">
          {{ t('search_bar.network_filter_label') + ' ' + headState.network }}
        </span>
        <fontAwesomeIcon class="arrow" :icon="faCaretDown" />
      </div>
    </BcSearchbarFilterButton>
    <div
      v-if="dropdownIsOpen"
      ref="dropdown"
      class="dropdown"
      :class="[barShape,colorTheme]"
      @keydown="(e) => {if (e.key === 'Escape') dropdownIsOpen = false}"
    >
      <div v-for="(line, i) of listInDropdown" :key="line.chainId" class="line" @click="oneOptionChanged(i)">
        <Checkbox v-model="line.selected" :binary="true" :input-id="String(line.chainId)" />
        <label :for="String(line.chainId)" class="label">
          {{ line.label }}
        </label>
        <IconNetwork
          :chain-id="line.chainId"
          :colored="true"
          :harmonize-perceived-size="true"
          :do-not-adapt-to-color-theme="colorTheme !== 'default'"
          class="icon"
        />
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
      width: 98px;
      .label {
        display: inline-flex;
        flex-grow: 1;
      }
      .arrow {
        position: relative;
        margin-top: auto;
        margin-bottom: auto;
      }
    }
  }

  .dropdown {
    position: absolute;
    display: block;
    box-sizing: border-box;
    z-index: 1024;
    border-radius: var(--padding);
    left: 0px;
    top: 24px;
    @media (pointer: coarse) {
      top: 32px;
    }
    padding: var(--padding);
    @include fonts.small_text_bold;

    &.default {
      background-color: var(--list-background);
      border: 1px solid var(--container-border-color);
      color: var(--text-color);
    }
    &.darkblue,
    &.lightblue {
      background-color: var(--searchbar-networkdropdown-bground-blue);
      border: 1px solid var(--searchbar-networkdropdown-border-blue);
      color: var(--light-black);
    }

    .line {
      position:relative;
      display: flex;
      width: 100%;
      white-space: nowrap;
      &:nth-child(n+1) {
        margin-top: var(--padding-small);
      }

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
        cursor: pointer;
        user-select: none;
        margin-top: auto;
        margin-bottom: auto;
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
