<script setup lang="ts">
import type { BcMainFooter } from '#build/components'
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
    <BcMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.content{
    padding: var(--padding-large);
}
.row {
    margin-bottom: var(--padding);
}
</style>
