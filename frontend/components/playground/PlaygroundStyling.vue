<script setup lang="ts">
const emptyModalVisibility = ref(false)
const headerPropModalVisibility = ref(false)
const slotModalVisibility = ref(false)
const loading = ref(true)

const toggleLoading = () => {
  loading.value = !loading.value
}

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

  <TabView>
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
    <TabPanel header="Spinner">
      <Button @click="toggleLoading">
        Toggle loading
      </Button>
      <div class="element_container">
        <BcLoadingSpinner :loading="loading" />
        <BcLoadingSpinner :loading="loading" size="small" style="color: lightblue;" />
        <BcLoadingSpinner :loading="loading" size="large" />
        <div class="box">
          <BcLoadingSpinner :loading="loading" alignment="center" />
        </div>
        <div class="box">
          <BcLoadingSpinner :loading="loading" size="full" />
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
.box{
  width: 200px;
  height: 200px;
  background-color: antiquewhite;
}
</style>
