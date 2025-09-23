<script setup lang="ts">
definePageMeta({
  layout: 'base',
  middleware: () => {
    return navigateTo('https://bitflyexplorergmbh.mintlify.app/api-reference/', { external: true })
  },
})
</script>

<template>
  <div>
    Page: API Docs
  </div>
</template>

<style scoped></style>
