<script setup lang="ts">
import { faTable } from '@fortawesome/pro-solid-svg-icons'
import { faChartColumn } from '@fortawesome/pro-regular-svg-icons'
import {
  IconAccount,
  IconSlotBlockProposal,
  IconValidator,
} from '#components'
import type { HashTabs } from '~/types/hashTabs'

const emptyModalVisibility = ref(false)
const headerPropModalVisibility = ref(false)
const slotModalVisibility = ref(false)
const isTable = ref<boolean>(true)
const isAttestation = ref<boolean>(true)

const loading = ref(true)

const toggleLoading = () => {
  loading.value = !loading.value
}

const selected = ref(true)

const completeList = [
  { value: 'attestation' },
  {
    component: IconSlotBlockProposal,
    value: 'proposal',
  },
  { value: 'sync' },
  {
    icon: faChartColumn,
    value: 'chart',
  },
]
const selectedList = ref<string[]>([
  'attestation',
  'proposal',
])

const selectedType = ref<string>('Validators')
const allTypes = [
  {
    component: IconAccount,
    text: 'Accounts',
    value: 'Accounts',
  },
  {
    component: IconValidator,
    text: 'Validators',
    value: 'Validators',
  },
]

const dropodownSelection = ref<string | undefined>()
const dropdownList = [
  {
    label: 'Yes',
    value: 'yes',
  },
  {
    label: 'No',
    value: 'no',
  },
  {
    label: 'Maybe we need a bigger label',
    value: 'maybe',
  },
]

const tabs: HashTabs = [
  {
    key: 'buttons',
    title: 'Buttons',
  },
  {
    key: 'hashes',
    title: 'Hashes',
  },
  {
    key: 'scroll',
    title: 'Scroll Box',
  },
  {
    key: 'input',
    title: 'Input',
  },
  {
    key: 'checkbox',
    title: 'Checkbox',
  },
  {
    key: 'toggle',
    title: 'Toggle',
  },
  {
    key: 'dropdown',
    title: 'Dropdown',
  },
  {
    key: 'spinner',
    title: 'Spinner',
  },
  {
    disabled: true,
    key: 'disabled',
    title: 'Disabled Tab',
  },
]
</script>

<template>
  <BcDialog v-model="emptyModalVisibility">
    <div class="element_container">
      <Button
        label="Close"
        @click="emptyModalVisibility = false"
      />
    </div>
  </BcDialog>
  <BcDialog
    v-model="headerPropModalVisibility"
    header="Text via Header Prop"
  >
    <div class="element_container">
      <Button
        label="Close"
        @click="headerPropModalVisibility = false"
      />
    </div>
  </BcDialog>
  <BcDialog
    v-model="slotModalVisibility"
    header="HeaderProp - Ignored as header slot wins"
  >
    <template #header>
      Utilizing the header slot for custom content
    </template>
    <div>
      Utilizing the default slot for custom content
      <br>
      <Button
        label="Close"
        @click="slotModalVisibility = false"
      />
    </div>

    <template #footer>
      Utilizing the footer slot for custom content
    </template>
  </BcDialog>

  <BcTabList
    :tabs default-tab="buttons"
  >
    <template #tab-panel-buttons>
      <div class="element_container">
        <Button> Text Button </Button>
        <Button
          label="Empty Modal"
          @click="emptyModalVisibility = true"
        />
        <Button
          label="Header Prop Modal"
          @click="headerPropModalVisibility = true"
        />
        <Button
          label="Slots Modal"
          @click="slotModalVisibility = true"
        />
        <Button>
          <BcLink to="/dashboard">
            Dashboard Link
          </BcLink>
        </Button>
        <Button disabled>
          Disabled
        </Button>
        <Button class="p-button-icon-only">
          <IconPlus
            alt="Plus icon"
            width="100%"
            height="100%"
          />
        </Button>
      </div>
    </template>

    <template #tab-panel-hashes>
      <PlaygroundHashes />
    </template>
    <template #tab-panel-scroll>
      <div class="scroll-box">
        <div>Scroll me</div>
      </div>
    </template>
    <template #tab-panel-input>
      <div class="element_container">
        <InputText placeholder="Input" />
        <InputText
          placeholder="Disabled Input"
          disabled
        />
      </div>
    </template>
    <template #tab-panel-checkbox>
      <div class="element_container">
        default checkbox:
        <Checkbox
          v-model="selected"
          :binary="true"
        /> disabled:
        <Checkbox disabled />
      </div>
    </template>
    <template #tab-panel-toggle>
      <h1>Multi Toggle</h1>
      <div class="element_container">
        <div>
          isTable: {{ isTable }}
          <BcIconToggle
            v-model="isTable"
            :true-icon="faTable"
            :false-icon="faChartColumn"
          />
        </div>

        <div>
          isAttestation: {{ isAttestation }}
          <BcIconToggle v-model="isAttestation">
            <template #trueIcon>
              <IconSlotAttestation />
            </template>

            <template #falseIcon>
              <IconSlotBlockProposal />
            </template>
          </BcIconToggle>
        </div>

        <div>
          Selected: {{ selected }}
          <BcToggleMultiBarButton
            v-model="selected"
            :icon="faTable"
          />
        </div>
        <div>
          <BcToggleMultiBar
            v-model="selectedList"
            :buttons="completeList"
            style="margin-right: 10px"
          >
            <template #attestation>
              <IconSlotAttestation />
            </template>

            <template #sync>
              <IconSlotSync />
            </template>
          </BcToggleMultiBar>
          Selected: {{ selectedList.join(", ") }}
        </div>
      </div>
      <h1>Single Toggle</h1>
      <div class="element_container">
        selectedType: {{ selectedType }}
        <BcToggleSingleBar
          v-model="selectedType"
          :buttons="allTypes"
          class="single_bar_container"
          layout="gaudy"
          :allow-deselect="true"
        />
      </div>
    </template>
    <template #tab-panel-dropdown>
      <div
        class="element_container"
        style="background-color: darkred; padding: 5px"
      >
        <BcDropdown
          v-model="dropodownSelection"
          :options="dropdownList"
          option-value="value"
          option-label="label"
          placeholder="for rock wtf this is a long placeholder"
          panel-style="max-width: 100px"
          style="max-width: 100px"
        />
        <BcDropdown
          v-model="dropodownSelection"
          :options="dropdownList"
          option-value="value"
          option-label="label"
          variant="table"
          placeholder="and roll"
          style="width: 200px"
        />
        Selected: {{ dropodownSelection }}
      </div>
    </template>
    <template #tab-panel-spinner>
      <Button @click="toggleLoading">
        Toggle loading
      </Button>
      <div class="element_container">
        <BcLoadingSpinner :loading />
        <BcLoadingSpinner
          :loading
          size="small"
          style="color: lightblue"
        />
        <BcLoadingSpinner
          :loading
          size="large"
        />
        <div class="box">
          <BcLoadingSpinner
            :loading
            alignment="center"
          />
        </div>
        <div class="box">
          <BcLoadingSpinner
            :loading
            size="full"
          />
        </div>
      </div>
    </template>
  </BcTabList>
</template>

<style lang="scss" scoped>
.element_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: var(--padding);
}

.box {
  width: 200px;
  height: 200px;
  background-color: antiquewhite;
}

.scroll-box {
  width: 100px;
  height: 100px;
  overflow: auto;
  div {
    background-color: grey;
    width: 200px;
    height: 200px;
  }
}

.single_bar_container {
  width: 600px;
}
</style>
