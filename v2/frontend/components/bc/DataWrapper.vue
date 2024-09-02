<script setup lang="ts">
// The DataWrapper is for loading the Data that used in the whole app.
// We can't load the data directly in the app.vue as this would conflict with some providers being initialized there.
const { getUser } = useUserStore()
const { tick } = useInterval(12)
const { refreshLatestState } = useLatestStateStore()
const {
  loadAvailableNetworks, setCurrentNetwork,
} = useNetworkStore()

await useAsyncData('latest_state', () => refreshLatestState(), { watch: [ tick ] })
await useAsyncData('get_user', () => getUser())
await useAsyncData('get-supported-networks', () => loadAvailableNetworks())

if (useRuntimeConfig().public.chainIdByDefault) {
  setCurrentNetwork(Number(useRuntimeConfig().public.chainIdByDefault))
}
</script>

<template>
  <slot />
</template>
