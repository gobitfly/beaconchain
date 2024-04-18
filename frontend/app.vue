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
useDateProvider()

</script>

<template>
  <div class="min-h-full">
    <BcDataWrapper>
      <NuxtPage />
      <DynamicDialog />
      <Toast />
    </BcDataWrapper>
  </div>
</template>

<style lang="scss">
</style>
