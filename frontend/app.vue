<script setup lang="ts">
import { generateUUID } from '~/utils/misc'

// Used for debugging purposes, might be removed or moved later
provide('app-uuid', { value: generateUUID() })
useHead(
  {
    script: [ {
      async: false,
      key: 'revive',
      src: '../js/revive.min.js',
    } ],
  },
  { mode: 'client' },
)
useWindowSizeProvider()
useBcToastProvider()
useDateProvider()
</script>

<template>
  <div class="min-h-full">
    <BcDataWrapper>
      <NuxtLoadingIndicator color="var(--primary-color)" />
      <NuxtPage />
      <DynamicDialog />
      <Toast />
    </BcDataWrapper>
  </div>
</template>

<style lang="scss"></style>
