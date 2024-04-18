<script setup lang="ts">
// The DataWrapper is for loading the Data that used in the whole app.
// We can't load the data directly in the app.vue as this would conflict with some providers being initialized there.
const { getUser } = useUserStore()

// TODO: load user on server once we fix SSR
await useAsyncData('get_user', () => getUser(), { server: false })
</script>
<template>
  <slot />
</template>
