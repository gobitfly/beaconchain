<script setup lang="ts">
const {
  is = 'span',
  size = 'md',
} = defineProps<{
  is?:
    | 'span',
  size?:
    | '2xl'
    | 'sm',
  variant?: 'secondary',
}>()
</script>

<template>
  <component
    :is
    class=""
    :class="[
      size === 'sm' && 'text-sm',
      size === '2xl' && 'text-2xl',
      variant === 'secondary' && 'dark:text-gray-400 text-gray-600',
    ]"
  >
    <slot />
  </component>
</template>
