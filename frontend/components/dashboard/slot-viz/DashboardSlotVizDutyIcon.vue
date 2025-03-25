<script setup lang="ts">
import type { SlotVizIcons } from '~/types/dashboard/slotViz'

interface Props {
  icon: SlotVizIcons,
}
const props = defineProps<Props>()
</script>

<template>
  <BcIcon
    v-if="props.icon === 'attestation'"
    name="file-signature"
  />
  <BcIcon
    v-else-if="props.icon === 'head_attestation'"
    name="attestaion-head"
  />
  <BcIcon
    v-else-if="props.icon === 'source_attestation'"
    name="attestaion-source"
  />
  <BcIcon
    v-else-if="props.icon === 'target_attestation'"
    name="attestaion-target"
  />
  <BcIcon
    v-else-if="props.icon === 'proposal'"
    name="cube"
  />
  <BcIcon
    v-else-if="props.icon === 'slashing'"
    name="user-slash"
  />
  <BcIcon
    v-else-if="props.icon === 'sync'"
    name="sync"
  />
</template>
