<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { type SlotVizData } from '~/types/dashboard/slotViz'
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
    <div class="row">
      latest epoch: {{ latest?.currentEpoch }}
    </div>

    <NuxtLink to="/" class="row">
      <Button>
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </NuxtLink>
    <div class="icon_holder">
      <div>
        <BcTooltip position="left" text="left tt">
          <IconSlotAttestation /> Attestation
        </BcTooltip>
      </div>
      <div>
        <BcTooltip position="right" text="right tt">
          <IconSlotHeadAttestation /> Head Attestation
        </BcTooltip>
      </div>
      <div>
        <BcTooltip position="top" text="top tt">
          <IconSlotSourceAttestation /> Source Attestation
        </BcTooltip>
      </div>
      <div>
        <BcTooltip position="bottom" text="bottom tt">
          <IconSlotTargetAttestation /> Target Attestation
        </BcTooltip>
      </div>
      <div>
        <IconSlotBlockProposal /> Block Proposal
      </div>
      <div>
        <IconSlotSlashing /> Slashing
      </div>
      <div>
        <IconSlotSync /> Slot Sync
      </div>
    </div>

    <div>
      <PlaygroundConversion />
    </div>
    <div class="icon_holder">
      <SlotVizViewer v-if="slotVizData" :data="slotVizData" />
    </div>
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
    <BcFooterMainFooter />
  </div>
</template>

<style lang="scss" scoped>

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
.icon_holder {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.content {
  padding: var(--padding-large);
}

.row {
  margin-bottom: var(--padding);
}

:deep(.bad-color){
  color: pink
}
</style>
