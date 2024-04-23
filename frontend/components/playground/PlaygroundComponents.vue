import type { ClientOnly } from '#build/components';
<template>
  <TabView lazy>
    <TabPanel header="Chart">
      <ClientOnly>
        <PlaygroundCharts />
      </ClientOnly>
    </TabPanel>
    <TabPanel header="Tags">
      <PlaygroundTags />
    </TabPanel>
    <TabPanel header="Duty Status">
      <PlaygroundDutyStatus />
    </TabPanel>
    <TabPanel header="Table">
      <PlaygroundTable />
    </TabPanel>
    <TabPanel header="Icons">
      <PlaygroundIcons />
    </TabPanel>
    <TabPanel header="Clipping in the middle">
      <PlaygroundMiddleEllipsisTests />
    </TabPanel>
    <TabPanel header="Search bars">
      <PlaygroundEmbeddedSearchbars />
    </TabPanel>
    <TabPanel header="Input">
      <PlaygroundInput />
    </TabPanel>
    <TabPanel header="Dialog">
      <PlaygroundDialog />
    </TabPanel>
  </TabView>
</template>
