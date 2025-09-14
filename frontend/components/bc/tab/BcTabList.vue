<script setup lang="ts">
import type { HashTabs } from '~/types/hashTabs'

interface Props {
  defaultTab: string,
  panelsClass?: string,
  tabs: HashTabs,
  useRouteHash?: boolean,
}
const props = defineProps<Props>()

const {
  activeTab,
} = useHashTabs(props.tabs, props.defaultTab, props.useRouteHash)
</script>

<template>
  <PvTabs
    v-model:value="activeTab"
    lazy
    scrollable
    class="dashboard-tab-view"
  >
    <PvTabList>
      <PvTab
        v-for="tab in tabs"
        :key="tab.key"
        :value="tab.key"
        :disabled="tab.disabled"
      >
        <BcTabHeader
          :header="tab.title"
          :icon="tab.icon"
        >
          <template #icon>
            <slot :name="`tab-header-icon-${tab.key}`" />
          </template>
        </BcTabHeader>
      </PvTab>
    </PvTabList>

    <PvTabPanels :class="panelsClass">
      <PvTabPanel
        v-for="tab in tabs"
        :key="tab.key"
        :value="tab.key"
      >
        <slot :name="`tab-panel-${tab.key}`">
          <component
            :is="tab.component"
            v-if="tab.component"
          />
          <div v-else-if="tab.placeholder">
            {{ tab.placeholder }}
          </div>
          <div v-else>
            tab-panel-{{ tab.key }}
          </div>
        </slot>
      </PvTabPanel>
    </PvTabPanels>
  </PvTabs>
</template>
