<template>
  <TabView>
    <TabPanel header="Buttons">
      <div class="element_container">
        <Button>
          Simple Text Button
        </Button>
        <Button label="Labeled" />
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
