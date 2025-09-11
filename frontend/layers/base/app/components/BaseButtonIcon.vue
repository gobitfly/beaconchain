<script setup lang="ts">
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const { size = 'md' } = defineProps<{
  name: IconName,
  size?: 'md',
  variant: 'secondary',
}>()
</script>

<template>
  <button
    class="flex rounded-full bg-linear-to-b disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'secondary' && 'from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90  ',
      size === 'md' && 'p-md',
    ]"
  >
    <BaseIcon
      :name
      class="w-6 h-6"
    />
  </button>
</template>

<style scoped></style>
