<script setup lang="ts" generic="T extends Tab[]">
export type Tab = {
  key: string,
  label: TranslationInput,
}
const props = defineProps<{
  classList?: {
    activeTab?: string,
    activeTabIndicator?: string,
    tab?: string,
    tablist?: string,
    tabpanel?: string,
  },
  defaultSelectedTab?: number,
  hasFocusableElement?: boolean,
  screenreaderTitle: TranslationInput,
  tabs: T,
}>()

const { t: $t } = useTranslation()

// otherwise syntax highlighting gets confused by typecasting 🫤
const ariaLabel = computed(() => $t(props.screenreaderTitle as string))

const idTab = useId()
const idTabPanel = useId()

const selectedTab = ref(props.defaultSelectedTab ?? 0)

const activeTabIndicators = useTemplateRef('activeTabIndicator')
const tabButtons = useTemplateRef('tab')

const moveIndicator = async (index: number) => {
  const activeTabIndicator = activeTabIndicators.value?.[0]
  const tabButton = tabButtons.value?.[index]
  if (!activeTabIndicator) return
  if (!tabButton) return

  const { x: initialX } = activeTabIndicator.getBoundingClientRect()
  const {
    width,
    x,
  } = tabButton.getBoundingClientRect()
  // this should rather have have been done via view transition api
  // but it currently lacks `firefox support`
  // and also there was a flickering issue with the activeTabIndicators height
  const animation = activeTabIndicator.animate([ {
    transform: `translateX(${x - initialX}px)`,
    width: `${width}px`,
  } ],
  {
    duration: 180,
  })
  return await animation.finished.then(() => {
    tabButton.appendChild(activeTabIndicator)
  })
}
const addActiveTabClassList = (index: number) => {
  const tabButton = tabButtons.value?.[index]
  if (!tabButton) return
  tabButton.classList.add(...(props.classList?.activeTab?.split(' ') ?? []))
}
const removeActiveTabClassList = (index: number) => {
  const tabButton = tabButtons.value?.[index]
  if (!tabButton) return
  tabButton.classList.remove(...(props.classList?.activeTab?.split(' ') ?? []))
}

const handleRight = () => {
  selectedTab.value = (selectedTab.value + 1) % (props.tabs.length)
}
const handleLeft = () => {
  selectedTab.value = (selectedTab.value - 1 + (props.tabs.length)) % (props.tabs.length)
}
const handleClick = async (index: number) => {
  selectedTab.value = index
}
watch(selectedTab, (newValue, oldValue) => {
  moveIndicator(newValue)
    .then(() => {
      removeActiveTabClassList(oldValue ?? 0)
    })
    .then(() => {
      addActiveTabClassList(newValue)
    })
    .then(() => {
      const nextTab = tabButtons.value?.[newValue]
      nextTab?.focus()
    })
}, { immediate: true })
</script>

<template>
  <div class="isolate">
    <section
      role="tablist"
      :aria-label
      :class="classList?.tablist"
      @keydown.right="handleRight"
      @keydown.left="handleLeft"
      @keydown.home.prevent="selectedTab = 0"
      @keydown.end.prevent="selectedTab = tabs.length - 1"
    >
      <button
        v-for="(tab, index) in tabs"
        :id="`${idTab}-${index}`"
        :key="tab.key"
        ref="tab"
        class="relative"
        :tabindex="selectedTab === index ? 0 : -1"
        type="button"
        role="tab"
        :aria-selected="selectedTab === index"
        :aria-controls="`${idTabPanel}-${index}`"
        :class="[classList?.tab, selectedTab === index && classList?.activeTab]"
        @click="handleClick(index)"
      >
        <span
          v-if="index === defaultSelectedTab"
          ref="activeTabIndicator"
          class="absolute inset-[0] activeTabIndicator z-0"
          aria-hidden="true"
          :class="classList?.activeTabIndicator"
        />
        <span
          :style="`view-transition-name: tabText-${index};`"
          class="tabText relative z-10"
        >
          {{ $t(tab.label as string) }}
        </span>
      </button>
    </section>
    <section>
      <template
        v-for="(tab, index) in tabs"
        :key="tab.key"
      >
        <article
          v-if="selectedTab === index"
          :id="`${idTabPanel}-${index}`"
          :class="classList?.tabpanel"
          tabindex="0"
          role="tabpanel"
          :aria-labelledby="`${idTab}-${index}`"
        >
          <slot :name="`tabpanel-${tab.key}`" />
        </article>
      </template>
    </section>
  </div>
</template>
