<script setup lang="ts">
import type { NuxtLinkProps } from '#app'
import type { IconName } from '#layers/base/app/components/BaseIcon.vue'

export type BaseNavigationItem = {
  icon: IconName,
  label: string,
  to: NuxtLinkProps['to'],
}

defineProps<BaseNavigationItem>()
</script>

<template>
  <NuxtLink
    class="flex items-center gap-md rounded-full border border-gray-200 bg-linear-to-b px-md py-sm text-sm font-semibold opacity-90 hover:opacity-95 dark:border-transparent dark:from-charcoal-600 dark:to-charcoal-700"
    :to
  >
    <BaseIcon :name="icon" />
    {{ label }}
  </NuxtLink>
</template>

<style scoped></style>
