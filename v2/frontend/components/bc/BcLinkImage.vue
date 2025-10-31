<script setup lang="ts">
import type { RouterLinkProps } from 'vue-router'

defineProps<{
  screenreaderText: string,
  to: RouterLinkProps['to'],
}>()
</script>

<template>
  <NuxtLink :to>
    <BcScreenreaderOnly>
      {{ screenreaderText }}
    </BcScreenreaderOnly>
    <slot />
  </NuxtLink>
</template>

<style scoped lang="scss"></style>
