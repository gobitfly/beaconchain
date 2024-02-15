<script setup lang="ts">
const modalVisibility = ref(false)

</script>

<template>
  <Dialog
    v-model:visible="modalVisibility"
    modal
    header="Dashboard &quot;Hetzner&quot;"
    :dismissable-mask="true"
    :closable="false"
    :draggable="false"
  >
    <Button type="button" label="Close" @click="modalVisibility = false" />
  </Dialog>

  <TabView>
    <TabPanel header="Buttons">
      <div class="element_container">
        <Button>
          Simple Text Button
        </Button>
        <Button label="Open Modal" @click="modalVisibility = true" />
        <Button>
          <NuxtLink to="/dashboard">
            Dashboard Link
          </NuxtLink>
        </Button>
        <Button :disabled="true">
          Disabled
        </Button>
        <Button class="p-button-icon-only">
          <IconPlus alt="Plus icon" />
        </Button>
      </div>
    </TabPanel>
    <TabPanel header="Input">
      <div class="element_container">
        <InputText placeholder="Input" />
        <InputText placeholder="Disabled Input" disabled />
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
