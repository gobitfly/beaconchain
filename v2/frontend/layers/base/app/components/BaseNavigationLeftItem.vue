<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

defineProps<BaseNavigationItem>()
</script>

<template>
  <NuxtLink
    class="font-semibold flex gap-md items-center p-lg opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300 rounded-xl"
    :to
  >
    <BaseIcon :name="icon" />
    {{ label }}
  </NuxtLink>
</template>

<style scoped></style>
