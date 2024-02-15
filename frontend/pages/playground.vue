<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { type SlotVizData } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
const { getLatestState } = useLatestStateStore()
await useAsyncData('latest_state', () => getLatestState())

const { latest } = storeToRefs(useLatestStateStore())

const slotVizData = ref<SlotVizData | null>(null)

await useAsyncData('test_slot_viz_data', async () => {
  const res = await $fetch<SlotVizData>('./mock/dashboard/slotViz.json')
  slotVizData.value = res
})

onMounted(async () => {
  const res = await $fetch<SlotVizData>('./mock/dashboard/slotViz.json')
  slotVizData.value = res
})

</script>

<template>
  <div class="content">
    <BcAdControl />
    <h1>Playground for testing UI components</h1>
    <NuxtLink to="/" class="row">
      <Button class="row">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </NuxtLink>
    <div class="row">
      Latest Epoch: {{ formatNumber(latest?.currentEpoch) }}
    </div>

    <TabView>
      <TabPanel header="Icons">
        <PlaygroundIcons />
      </TabPanel>
      <TabPanel header="Conversion">
        <PlaygroundConversion />
      </TabPanel>
      <TabPanel header="Slot Viz">
        <SlotVizViewer v-if="slotVizData" :data="slotVizData" />
      </TabPanel>
      <TabPanel header="Ads">
        <div class="ad_test_container">
          No blue box should be left here
          <div id="replace_me" class="ad_test">
            Ok come on and replace me
          </div>
        </div>
        <div class="ad_test_container">
          Ad should be within the box
          <div id="inside_me" class="ad_test">
            Should be iniside
          </div>
        </div>
        <div class="ad_test_container">
          Ad should be after this text, but before the box
          <div id="before_me" class="ad_test">
            Should come before me
          </div>
        </div>
        <div class="ad_test_container">
          <div id="after_me" class="ad_test">
            Should come after me
          </div>
          Ad should be before this text, but after the box
        </div>
        <div class="ad_test_container">
          Ad should be after this text, but before the box
          <div id="around_me" class="ad_test">
            Should come around me
          </div>
          Ad should be before this text, but after the box
        </div>
      </TabPanel>
      <TabPanel :disabled="true" header="Disabled" />
    </TabView>

    <BcFooterMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.icon_holder {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.ad_test_container{
  color: red;
  border: 1px solid red;
  background-color: aqua;
  padding: 10px;
}
.ad_test{
  color: pink;
  border: 1px solid pink;
  background-color: darkblue;
  padding: 10px;
}

.content {
  padding: var(--padding-large);
}

.row {
  margin-bottom: var(--padding);
}
</style>
