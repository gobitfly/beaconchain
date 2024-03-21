<script setup lang="ts">
useHead({
  script: [
    {
      key: 'revive',
      src: '../js/revive.min.js',
      async: false
    }
  ]
}, { mode: 'client' })
useWindowSizeProvider()
useBcToastProvider()

</script>

<template>
  <div class="min-h-full">
    <NuxtPage />
    <DynamicDialog />
    <Toast />
  </div>
</template>

<style lang="scss">
</style>
