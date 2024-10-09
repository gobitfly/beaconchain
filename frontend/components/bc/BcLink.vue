<script setup lang="ts">
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
