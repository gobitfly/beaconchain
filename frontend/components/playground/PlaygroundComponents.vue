import PlaygroundSearchbarValidators from './PlaygroundSearchbarValidators.vue';

import PlaygroundSearchbarAccounts from './PlaygroundSearchbarAccounts.vue';
<template>
  <TabView lazy>
    <TabPanel header="Table">
      <PlaygroundTable />
    </TabPanel>
    <TabPanel header="Chart">
      <PlaygroundCharts />
    </TabPanel>
    <TabPanel header="Icons">
      <PlaygroundIcons />
    </TabPanel>
    <TabPanel header="Search bars">
      <PlaygroundEmbeddedSearchbars />
    </TabPanel>
    <TabPanel header="Dialog">
      <PlaygroundDialog />
    </TabPanel>
  </TabView>
</template>
