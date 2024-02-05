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
