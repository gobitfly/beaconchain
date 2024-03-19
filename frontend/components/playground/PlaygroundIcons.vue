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
    <h2>Network icons</h2>
    <div>Monochromous:</div>
    <IconNetwork
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="false"
      :harmonize-perceived-size="false"
      class="icon-size"
    />
    <div>Colored:</div>
    <IconNetwork
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="true"
      :harmonize-perceived-size="false"
      class="icon-size"
    />
    <div>Harmonized sizes (for example Ethereum looks less skinny compared to the others):</div>
    <IconNetwork
      v-for="id of getListOfChainIDs(false)"
      :key="id"
      :chain-id="id as ChainIDs"
      :colored="true"
      :harmonize-perceived-size="true"
      class="icon-size"
    />
  </div>
  <div>
    <h2>Loading issues:</h2>
    Example when props `inline` is true: "Something went wrong <IconErrorFace :inline="true" /> Please try again later."
    <br>
    Example for a 404 page (not inline):
    <div class="container">
      <br>
      <div>Error 404</div>
      <br>
      <IconErrorFace width="30%" />
    </div>
  </div>
  <div>
    <h2>Data types:</h2>
    <IconDatatypeTokenIcon class="icon-size" />
    <IconDatatypeNftIcon class="icon-size" />
    <IconDatatypeEpochIcon class="icon-size" />
    <IconDatatypeBlockIcon class="icon-size" />
    <IconDatatypeTransactionIcon class="icon-size" />
    <IconDatatypeBatchIcon class="icon-size" />
    <IconDatatypeContractIcon class="icon-size" />
    <IconDatatypeAccountIcon class="icon-size" />
    <IconDatatypeEnsIcon class="icon-size" style="color:red" />
    <IconDatatypeEnsIcon class="icon-size" style="color:blue" />
    <IconDatatypeGraffitiIcon class="icon-size" />
    <IconDatatypeValidatorIcon class="icon-size" />
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

.icon-size {
  height: 40px;
  width: 40px;
}

.container {
  width: 500px;
  height: 300px;
  text-align: center;
}
</style>
