<script setup lang="ts">
definePageMeta({
  layout: 'base',
  middleware: () => {
    return navigateTo(externalLink.beaconchain.apiDocs, { external: true })
  },
})
</script>

<template>
  <div>
    Page: API Docs
  </div>
</template>

<style scoped></style>
