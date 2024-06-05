<script setup lang="ts">
import { generateUUID } from '~/utils/misc'
import { useNetwork } from '~/composables/useNetwork'

// Used for debugging purposes, might be removed or moved later
provide('app-uuid', { value: generateUUID() })
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

const { setCurrentNetwork } = useNetwork()
if (useRuntimeConfig().public.chainIdByDefault) {
  setCurrentNetwork(Number(useRuntimeConfig().public.chainIdByDefault))
}

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
