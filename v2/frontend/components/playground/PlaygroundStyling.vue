<script setup lang="ts">
const dashboardModalVisibility = ref(false)
const otherModalVisibility = ref(false)
</script>

<template>
  <BcDialog v-model="dashboardModalVisibility" header="Dashboard &quot;Hetzner&quot;">
    <div class="element_container">
      <Button label="Close" @click="dashboardModalVisibility = false" />
    </div>
  </BcDialog>
  <BcDialog v-model="otherModalVisibility" header="This is another modal!">
    <div class="element_container">
      <Button label="Close this other modal!" @click="otherModalVisibility = false" />
    </div>
  </BcDialog>

  <TabView>
    <TabPanel header="Buttons">
      <div class="element_container">
        <Button>
          Simple Text Button
        </Button>
        <Button label="Open Modal" @click="dashboardModalVisibility = true" />
        <Button label="Open Other Modal" @click="otherModalVisibility = true" />
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
