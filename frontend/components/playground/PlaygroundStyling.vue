<template>
  <TabView>
    <TabPanel header="Buttons">
      <div class="button_container">
        <Button>
          Simple Text Button
        </Button>
        <Button label="Labeled" />
        <NuxtLink to="/dashboard">
          <Button label="Dashboard Link" />
        </NuxtLink>
        <Button :disabled="true">
          Disabled
        </Button>
        <Button class="p-button-icon-only">
          <IconPlus alt="Plus icon" />
        </Button>
      </div>
    </TabPanel>
    <TabPanel header="Input">
      <InputText label="Input" />
    </TabPanel>
    <TabPanel :disabled="true" header="Disabled" />
  </TabView>
</template>

<style lang="scss" scoped>
.button_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
