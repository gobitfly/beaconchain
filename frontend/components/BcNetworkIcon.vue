<script setup lang="ts">
import type { ChainId } from '~/types/network'

const { id } = defineProps<{
  id: ChainId,
}>()

const { getNetworkName } = useNetwork()

const networkName = computed(() => getNetworkName(id))
</script>

<template>
  <BcIcon
    v-if="networkName === 'Ethereum'"
    name="crypto-eth"
  />
  <BcIcon
    v-if="networkName === 'Gnosis'"
    name="crypto-gno"
  />
</template>
