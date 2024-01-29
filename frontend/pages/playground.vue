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
        <IconSlotAttestation /> Attestation
      </div>
      <div>
        <IconSlotHeadAttestation /> Head Attestation
      </div>
      <div>
        <IconSlotSourceAttestation /> Source Attestation
      </div>
      <div>
        <IconSlotTargetAttestation /> Target Attestation
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
    <div class="icon_holder">
      <SlotVizViewer v-if="slotVizData" :data="slotVizData" />
      <div>
        <SlotVizTile :data="{ state: 'scheduled', id: 1 }" /> Sceduled
      </div>
      <div>
        <SlotVizTile :data="{ state: 'scheduled', id: 1, duties: [{ type: 'proposal', pendingCount: 1, validator: 1234 }]}" /> Proposer duty
      </div>
      <div>
        <SlotVizTile :data="{ state: 'scheduled', id: 1, duties: [{ type: 'attestation', pendingCount: 3 }]}" /> Proposer duty
      </div>
      <div>
        <SlotVizTile :data="{ state: 'missed', id: 2 }" /> Missed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 3 }" /> Proposed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 4, duties: [{ type: 'proposal', failedCount: 1, failedEarnings: '11200000000000000', validator: 1234 }] }" /> Proposed, validator missed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 5, duties: [{ type: 'proposal', successCount: 1, successEarning: '11200000000000000', validator: 1234 }] }" /> Proposed, validator proposed
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 6, duties:[{type:'proposal', failedCount:1, failedEarnings: '11200000000000000', successCount: 1, successEarning: '11200000000000000', validator: 1234}]}" /> Proposed,
        validator mixed
      </div>
      <div>
        <SlotVizTile
          :data="{state:'proposed', id: 7, duties:[{type:'proposal', successCount: 1, validator: 1234, successEarning: '11200000000000000'},{type:'attestation', successCount: 3, successEarning: '11200000000000000'}]}"
        />
        Proposed, validator 2 icons
      </div>
      <div>
        <SlotVizTile
          :data="{state:'proposed', id: 8, duties:[{type:'proposal', failedCount:1, failedEarnings: '11200000000000000', validator: 1234},{type:'attestation', successCount: 4, successEarning: '11200000000000000'},{type:'slashing', successCount: 1, successEarning: '11200000000000000'}]}"
        />
        Proposed, 3 icons
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 9, duties:[{type:'slashing', successCount: 1, successEarning: '11200000000000000'}]}" /> Proposed, slash
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 10, duties:[{type:'sync', successCount: 1, successEarning: '11200000000000000'}]}" /> Proposed, slash
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 11, duties:[{type:'attestation', successCount: 1, successEarning: '11200000000000000'}]}" /> Proposed, validator
        pending
      </div>
    </div>
    <BcMainFooter />
  </div>
</template>

<style lang="scss" scoped>.icon_holder {
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
}</style>
