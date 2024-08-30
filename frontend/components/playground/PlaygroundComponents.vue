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

const tabs: HashTabs = [
  {
    component: PlaygroundDialog,
    key: 'dialog',
    title: 'Dialog',
  },
  {
    component: PlaygroundAccordeon,
    key: 'accordeon',
    title: 'Accordeon',
  },
  {
    component: PlaygroundMiddleEllipsisTests,
    key: 'middle',
    title: 'Middle Ellipsis',
  },
  {
    component: PlaygroundTags,
    key: 'tags',
    title: 'Tags',
  },
  {
    component: PlaygroundDutyStatus,
    key: 'duty',
    title: 'Duty Status',
  },
  {
    component: PlaygroundTable,
    key: 'table',
    title: 'Table',
  },
  {
    component: PlaygroundIcons,
    key: 'icons',
    title: 'Icons',
  },
  {
    component: PlaygroundInput,
    key: 'input',
    title: 'Input',
  },
]
</script>

<template>
  <BcTabList :tabs default-tab="icons" />
</template>
