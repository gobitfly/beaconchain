<script setup lang="ts">
// The DataWrapper is for loading the Data that used in the whole app.
// We can't load the data directly in the app.vue as this would conflict with some providers being initialized there.
const {
  getUser,
  isLoggedIn,
} = useUserStore()
const { networkInfo } = useNetwork()
const { secondsPerSlot } = networkInfo.value
const { counter } = useInterval(secondsPerSlot)
const { refreshLatestState } = useLatestStateStore()

await useAsyncData('latest_state', () => refreshLatestState(), {
  immediate: true,
  watch: [ counter ],
})
if (isLoggedIn) {
  await useAsyncData('get_user', () => getUser())
}
</script>

<template>
  <slot />
</template>
