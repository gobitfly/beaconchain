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
    <span v-for="id of getListOfChainIDs(false)" :key="id" class="network-icon-cont">
      <IconNetworkIcons :chain-id="id as ChainIDs" :colored="false" :harmonize-perceived-size="false" class="network-icon" />
    </span>
    <br>Colored:
    <span v-for="id of getListOfChainIDs(false)" :key="id" class="network-icon-cont">
      <IconNetworkIcons :chain-id="id as ChainIDs" :colored="true" :harmonize-perceived-size="false" class="network-icon" />
    </span>
    <br>Harmonious:
    <span v-for="id of getListOfChainIDs(false)" :key="id" class="network-icon-cont">
      <IconNetworkIcons :chain-id="id as ChainIDs" :colored="true" :harmonize-perceived-size="true" class="network-icon" />
    </span>
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

.network-icon-cont {
 /* display: flex;
  width: 300px;
  height: 300px;*/
}
.network-icon {
  display: flex;
  width: 50px;
  height: 50px;
}
</style>
