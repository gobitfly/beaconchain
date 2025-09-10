<script setup lang="ts">
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  size = 'md',
  variant = 'primary',
} = defineProps<{
  leadingIcon?: IconName,
  size?: 'md' | 'xl',
  trailingIcon?: IconName,
  variant?: 'branded' | 'primary' | 'secondary',
}>()
</script>

<template>
  <button
    class="flex items-center bg-linear-to-b rounded-full font-semibold disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'primary' && 'from-gray-700 to-gray-900 text-white dark:from-gray-100 dark:to-gray-300 dark:text-black opacity-90 hover:opacity-95',
      variant === 'secondary' && 'from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95',
      variant === 'branded' && 'from-brand-500 to-brand-700 text-white hover:opacity-90',
      size === 'md' && 'text-sm py-md px-sm',
      size === 'xl' && 'text-md py-xl px-3xl',
    ]"
  >
    <LazyBaseIcon
      v-if="leadingIcon"
      :name="leadingIcon"
      class=""
    />
    <span class="px-lg">
      <slot />
    </span>
    <LazyBaseIcon
      v-if="trailingIcon"
      :name="trailingIcon"
      class=""
    />
  </button>
</template>
