<script setup lang="ts">
definePageMeta({
  layout: 'base',
  middleware: () => {
    return navigateTo('/dashboard')
  },
})
</script>

<template>
  <div>
    Page: stakinghub
  </div>
</template>

<style scoped></style>
