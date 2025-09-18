<script setup lang="ts">
// The DataWrapper is for loading the Data that used in the whole app.
// We can't load the data directly in the app.vue as this would conflict with some providers being initialized there.
const { networkInfo } = useNetworkStore()
const { secondsPerSlot } = networkInfo.value
const { counter } = useInterval(secondsPerSlot)
const { refreshLatestState } = useLatestStateStore()

await useAsyncData('latest_state', () => refreshLatestState(), {
  watch: [ counter ],
})
</script>

<template>
  <slot />
</template>
