<script setup lang="ts">
const {
  size = 'sm',
  variant = 'neutral',
} = defineProps<{
  isSelected: boolean,
  size?: 'sm',
  variant?: 'neutral',
}>()
</script>

<template>
  <button
    class="text-nowrap border-1 font-semibold focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'neutral' && 'bg-gray-100 dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 aria-pressed:bg-gray-200 dark:aria-pressed:bg-gray-700 aria-pressed:border-gray-300 dark:aria-pressed:border-gray-700 aria-pressed:text-black dark:aria-pressed:text-white aria-pressed:shadow-none',
      size === 'sm' && 'px-md py-xs rounded-md ',
    ]"
    :aria-pressed="isSelected"
  >
    <slot />
  </button>
</template>
