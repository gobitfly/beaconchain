<script setup lang="ts">
import { getListOfChainIDs, ChainIDs } from '~/types/networks'
</script>

<template>
  <div id="iconHolder" class="icon_holder">
    <div>
      <BcTooltip position="left" text="left tt" scroll-container="#iconHolder">
        <IconSlotAttestation /> Attestation
      </BcTooltip>
    </div>
    <div>
      <BcTooltip position="right" text="right tt" scroll-container="#iconHolder">
        <IconSlotHeadAttestation /> Head Attestation
      </BcTooltip>
    </div>
    <div>
      <BcTooltip position="top" text="top tt" scroll-container="#iconHolder">
        <IconSlotSourceAttestation /> Source Attestation
      </BcTooltip>
    </div>
    <div>
      <BcTooltip position="bottom" text="bottom tt" scroll-container="#iconHolder">
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
    Monochromous:
    <IconNetworkIcons
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="false"
      :harmonize-perceived-size="false"
      class="network-icon"
    />
    <br>Colored:
    <IconNetworkIcons
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="true"
      :harmonize-perceived-size="false"
      class="network-icon"
    />
    <br>Harmonious:
    <IconNetworkIcons
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="true"
      :harmonize-perceived-size="true"
      class="network-icon"
    />
  </div>
</template>

<style lang="scss" scoped>
.icon_holder {
  margin: 10px;
  display: flex;
  flex-direction: column;
  overflow: auto;
  gap: var(--padding);
  height: 100px;
  width: 200px;
}

.network-icon {
  height: 40px;
  width: 40px;
}
</style>
