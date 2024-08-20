<script setup lang="ts">
import {
  PlaygroundAccordeon,
  PlaygroundDialog,
  PlaygroundDutyStatus,
  PlaygroundIcons,
  PlaygroundInput,
  PlaygroundMiddleEllipsisTests,
  PlaygroundTable,
  PlaygroundTags,
} from '#components'
import type { HashTabs } from '~/types/hashTabs'

const tabs: HashTabs = {
  accordeon: {
    component: PlaygroundAccordeon,
    index: '1',
    title: 'Accordeon',
  },
  dialog: {
    component: PlaygroundDialog,
    index: '0',
    title: 'Dialog',
  },
  duty: {
    component: PlaygroundDutyStatus,
    index: '4',
    title: 'Duty Status',
  },
  icons: {
    component: PlaygroundIcons,
    index: '6',
    title: 'Icons',
  },
  input: {
    component: PlaygroundInput,
    index: '7',
    title: 'Input',
  },
  middle: {
    component: PlaygroundMiddleEllipsisTests,
    index: '2',
    title: 'Middle Ellipsis',
  },
  table: {
    component: PlaygroundTable,
    index: '5',
    title: 'Table',
  },
  tags: {
    component: PlaygroundTags,
    index: '3',
    title: 'Tags',
  },
}
</script>

<template>
  <BcTabList :tabs default-tab="icons" />
</template>
