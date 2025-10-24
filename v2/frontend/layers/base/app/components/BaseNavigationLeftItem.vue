<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

defineProps<BaseNavigationItem>()
</script>

<template>
  <NuxtLink
    class="flex items-center gap-md rounded-xl p-lg text-sm font-semibold opacity-90 hover:opacity-95"
    :to
  >
    <BaseIcon :name="icon" />
    {{ label }}
  </NuxtLink>
</template>

<style scoped></style>
