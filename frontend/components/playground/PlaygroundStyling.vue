<script setup lang="ts">
import {
  faTable
} from '@fortawesome/pro-solid-svg-icons'
import {
  faChartColumn
} from '@fortawesome/pro-regular-svg-icons'
const emptyModalVisibility = ref(false)
const headerPropModalVisibility = ref(false)
const slotModalVisibility = ref(false)
const isTable = ref<boolean>(true)
const isAttestation = ref<boolean>(true)

</script>

<template>
  <BcDialog v-model="emptyModalVisibility">
    <div class="element_container">
      <Button label="Close" @click="emptyModalVisibility = false" />
    </div>
  </BcDialog>
  <BcDialog v-model="headerPropModalVisibility" header="Text via Header Prop">
    <div class="element_container">
      <Button label="Close" @click="headerPropModalVisibility = false" />
    </div>
  </BcDialog>
  <BcDialog v-model="slotModalVisibility" header="HeaderProp - Ignored as header slot wins">
    <template #header>
      Utilizing the header slot for custom content
    </template>
    <div>
      Utilizing the default slot for custom content
      <br>
      <Button label="Close" @click="slotModalVisibility = false" />
    </div>
    <template #footer>
      Utilizing the footer slot for custom content
    </template>
  </BcDialog>

  <TabView lazy>
    <TabPanel header="Buttons">
      <div class="element_container">
        <Button>
          Text Button
        </Button>
        <Button label="Empty Modal" @click="emptyModalVisibility = true" />
        <Button label="Header Prop Modal" @click="headerPropModalVisibility = true" />
        <Button label="Slots Modal" @click="slotModalVisibility = true" />
        <Button>
          <NuxtLink to="/dashboard">
            Dashboard Link
          </NuxtLink>
        </Button>
        <Button :disabled="true">
          Disabled
        </Button>
        <Button class="p-button-icon-only">
          <IconPlus alt="Plus icon" width="100%" height="100%" />
        </Button>
      </div>
    </TabPanel>
    <TabPanel header="Input">
      <div class="element_container">
        <InputText placeholder="Input" />
        <InputText placeholder="Disabled Input" disabled />
      </div>
    </TabPanel>
    <TabPanel header="Toggle">
      <div class="element_container">
        <div>isTable: {{ isTable }} <BcIconToggle v-model="isTable" :true-icon="faTable" :false-icon="faChartColumn" /></div>

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
      </div>
    </TabPanel>
    <TabPanel :disabled="true" header="Disabled Tab" />
  </TabView>
</template>

<style lang="scss" scoped>
.element_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
