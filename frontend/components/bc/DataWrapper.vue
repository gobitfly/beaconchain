<script setup lang="ts">
// The DataWrapper is for loading the Data that used in the whole app.
// We can't load the data directly in the app.vue as this would conflict with some providers being initialized there.
const {
  getUser,
  isLoggedIn,
} = useUserStore()
const { tick } = useInterval(12)
const { refreshLatestState } = useLatestStateStore()
const {
  setCurrentNetwork,
} = useNetworkStore()

await useAsyncData('latest_state', () => refreshLatestState(), {
  immediate: true,
  watch: [ tick ],
})
if (isLoggedIn) {
  await useAsyncData('get_user', () => getUser())
}

const { chainIdByDefault } = useRuntimeConfig().public
if (chainIdByDefault) {
  setCurrentNetwork(Number(chainIdByDefault))
}
</script>

<template>
  <slot />
</template>
