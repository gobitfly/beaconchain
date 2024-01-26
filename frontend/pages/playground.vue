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
      <div><IconSlotAttestation /> Attestation</div>
      <div><IconSlotHeadAttestation /> Head Attestation</div>
      <div><IconSlotSourceAttestation /> Source Attestation</div>
      <div><IconSlotTargetAttestation /> Target Attestation</div>
      <div><IconSlotBlockProposal /> Block Proposal</div>
      <div><IconSlotSlashing /> Slashing</div>
      <div><IconSlotSync /> Slot Sync</div>
    </div>
    <BcMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.icon_holder{
  margin: 10px;
  display: flex;
  flex-direction: column;
}
.content{
    padding: var(--padding-large);
}
.row {
    margin-bottom: var(--padding);
}
</style>
