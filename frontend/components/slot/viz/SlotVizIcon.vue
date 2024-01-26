<script setup lang="ts">
import { type SlotVizIcons } from '~/types/dashboard/slotViz'

interface Props {
  icon: SlotVizIcons
}
const props = defineProps<Props>()

</script>
<template>
  <IconSlotAttestation v-if="props.icon === 'attestation'" />
  <IconSlotHeadAttestation v-else-if="props.icon === 'head_attestation'" />
  <IconSlotSourceAttestation v-else-if="props.icon === 'source_attestation'" />
  <IconSlotTargetAttestation v-else-if="props.icon === 'target_attestation'" />
  <IconSlotBlockProposal v-else-if="props.icon === 'block_proposal'" />
  <IconSlotSlashing v-else-if="props.icon === 'slashing'" />
  <IconSlotSync v-else-if="props.icon === 'sync'" />
</template>
