<script setup lang="ts">
import { useLatestState } from '~/stores/latestState'

const { latest, getLatestState } = useLatestState()
const { data: latestTest } = await useAsyncData('latest_state', () => getLatestState())
</script>
<template>
  <div>
    test: {{ latestTest?.currentEpoch }}
    Current Epoch: {{ latest?.currentEpoch }}
    <NuxtLink to="/login" class="test-link">
      Login
    </NuxtLink>
  </div>
</template>

<style lang="scss" scoped>
.test-link{
  background-color: var(--text-color);
  color: var(--background-color);
}
</style>
