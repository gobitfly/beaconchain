<script setup lang="ts">
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

export type HashTab = {
  component?: Component,
  disabled?: boolean,
  icon?: Icon,
  key: string,
  placeholder?: string,
  title?: string,
}

const props = defineProps<{
  defaultTab: string,
  panelsClass?: string,
  queryParameterKey?: string,
  tabs: HashTab[],
}>()

const router = useRouter()
const route = useRoute()
const currentTabQueryValue = computed(() => {
  const queryValue = route.query[props.queryParameterKey ?? 'tab']
  // if query parameters are added more than once
  if (Array.isArray(queryValue)) return queryValue[0]
  return queryValue
})
const activeTab = ref(currentTabQueryValue.value ?? props.defaultTab)
const onUpdateValue = () => {
  if (!props.queryParameterKey) return
  router.push({
    query: {
      ...route.query,
      [props.queryParameterKey]: activeTab.value,
    },
  })
}
</script>

<template>
  <Tabs
    v-model:value="activeTab"
    lazy
    scrollable
    class="dashboard-tab-view"
    @update:value="onUpdateValue"
  >
    <TabList>
      <Tab
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
      </Tab>
    </TabList>

    <TabPanels :class="panelsClass">
      <TabPanel
        v-for="tab in tabs"
        :key="tab.key"
        class="tab-panel"
        :value="tab.key"
      >
        <slot
          :name="`tab-panel-${tab.key}`"
          v-bind="{ isActive: activeTab === tab.key }"
        >
          <component
            :is="tab.component"
            v-if="tab.component"
          />
          <div v-else-if="tab.placeholder">
            {{ tab.placeholder }}
          </div>
          <slot
            v-else
            name="empty"
          >
            tab-panel-{{ tab.key }}
          </slot>
        </slot>
      </TabPanel>
    </TabPanels>
  </Tabs>
</template>

<style lang="scss" scoped>
:deep(.tab-panel) {
  min-height: inherit;
  display: grid;
}
</style>
