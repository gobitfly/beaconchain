<template>
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
</template>

<style lang="scss" scoped>
.icon_holder {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
