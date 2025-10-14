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
    class="font-semibold flex gap-md items-center py-sm px-md bg-linear-to-b rounded-full border border-gray-200 dark:border-transparent dark:from-charcoal-600 dark:to-charcoal-700 opacity-90 hover:opacity-95 text-sm"
    :to
  >
    <BaseIcon :name="icon" />
    {{ label }}
  </NuxtLink>
</template>

<style scoped></style>
