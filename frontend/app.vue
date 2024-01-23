<script setup lang="ts">
import { useLatestState } from '~/stores/useLatestState'
const { getLatestState } = useLatestState()
await useAsyncData('latest_state', () => getLatestState())

</script>

<template>
  <div class="min-h-full">
    <NuxtPage />
  </div>
</template>

<style lang="scss">
</style>
