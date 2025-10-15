<script setup lang="ts">
definePageMeta({
  layout: 'base',
  middleware: () => {
    return navigateTo('https://docs.beaconcha.in ', { external: true })
  },
})
</script>

<template>
  <div>
    Page: API Docs
  </div>
</template>

<style scoped></style>
