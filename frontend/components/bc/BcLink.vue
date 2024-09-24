<script setup lang="ts">
// import type { RouteLocationRaw } from 'vue-router'
import type { NuxtLinkProps } from '#app'

defineProps<{
  to: NuxtLinkProps['to'],
}>()
</script>

<template>
  <NuxtLink
    :to
    :no-prefetch="true"
  >
    <template #default>
      <slot />
    </template>
  </NuxtLink>
</template>
