<template>
  <TabView>
    <TabPanel header="Buttons">
      <Button>
        Simple Button
      </Button>
    </TabPanel>
    <TabPanel :disabled="true" header="Disabled" />
  </TabView>
</template>

<style lang="scss" scoped>
</style>
