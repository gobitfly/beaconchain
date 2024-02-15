<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
useHead({
  script: [
    {
      key: 'revive',
      src: 'js/revive.min.js',
      async: false
    }
  ]
}, { mode: 'client' })
const { getLatestState } = useLatestStateStore()
await useAsyncData('latest_state', () => getLatestState())

</script>

<template>
  <div class="min-h-full">
    <NuxtPage />
  </div>
</template>

<style lang="scss">
</style>
