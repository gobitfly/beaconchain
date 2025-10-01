<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

defineProps<BaseNavigationItem>()
</script>

<template>
  <NuxtLink
    class="font-semibold flex gap-md items-center p-lg opacity-90 hover:opacity-95 text-sm rounded-xl"
    :to
  >
    <BaseIcon :name="icon" />
    {{ label }}
  </NuxtLink>
</template>

<style scoped></style>
