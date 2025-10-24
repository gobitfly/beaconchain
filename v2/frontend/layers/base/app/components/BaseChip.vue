<script setup lang="ts">
import type { IconName } from '#layers/base/app/components/BaseIcon.vue'

const {
  size = 'sm',
  variant = 'neutral',
} = defineProps<{
  icon?: IconName,
  isSelected: boolean,
  size?: 'sm',
  variant?: 'neutral',
}>()
</script>

<template>
  <button
    class="border font-semibold text-nowrap focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    type="button"
    :class="[
      variant === 'neutral' && 'border-gray-200 bg-gray-100 text-gray-600 aria-pressed:border-gray-300 aria-pressed:bg-gray-200 aria-pressed:text-black aria-pressed:shadow-none dark:border-gray-800 dark:bg-gray-900 dark:text-gray-400 dark:aria-pressed:border-gray-700 dark:aria-pressed:bg-gray-700 dark:aria-pressed:text-white',
      size === 'sm' && 'rounded-md px-md py-xs ',
      icon && 'flex items-center gap-md',
    ]"
    :aria-pressed="isSelected"
  >
    <BaseIcon
      v-if="icon"
      :name="icon"
    />
    <slot />
  </button>
</template>
