<script setup lang="ts">
import type { NuxtLinkProps } from '#app'

const { to } = defineProps<{
  to: NuxtLinkProps['to'],
}>()
</script>

<template>
  <NuxtLink
    :to
  >
    <template #default>
      <slot />
    </template>
  </NuxtLink>
</template>
