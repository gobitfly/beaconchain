<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
const { getLatestState } = useLatestStateStore()
await useAsyncData('latest_state', () => getLatestState())

const { latest } = storeToRefs(useLatestStateStore())

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
      <div>
        <SlotVizTile :data="{ state: 'scheduled', id: 1 }" /> Sceduled
      </div>
      <div>
        <SlotVizTile :data="{ state: 'missed', id: 2 }" /> Missed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 3 }" /> Proposed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 4, duties: [{ type: 'propsal', failedCount: 1 }] }" /> Proposed, validator missed
      </div>
      <div>
        <SlotVizTile :data="{ state: 'proposed', id: 5, duties: [{ type: 'propsal', successCount: 1 }] }" /> Proposed, validator proposed
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 6, duties:[{type:'propsal', failedCount:1, successCount: 1}]}" /> Proposed,
        validator mixed
      </div>
      <div>
        <SlotVizTile
          :data="{state:'proposed', id: 7, duties:[{type:'propsal', successCount: 1},{type:'attestation', successCount: 1}]}"
        />
        Proposed, validator 2 icons
      </div>
      <div>
        <SlotVizTile
          :data="{state:'proposed', id: 8, duties:[{type:'propsal', failedCount:1},{type:'attestation', successCount: 1},{type:'slashing', successCount: 1}]}"
        />
        Proposed, 3 icons
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 9, duties:[{type:'slashing', successCount: 1}]}" /> Proposed, slash
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 10, duties:[{type:'sync', successCount: 1}]}" /> Proposed, slash
      </div>
      <div>
        <SlotVizTile :data="{state:'proposed', id: 11, duties:[{type:'attestation', pendingCount:1}]}" /> Proposed, validator
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
