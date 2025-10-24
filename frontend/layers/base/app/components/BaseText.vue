<script setup lang="ts">
const {
  fontWeight = 'normal',
  is = 'span',
  size = 'md',
} = defineProps<{
  fontWeight?: 'bold' | 'normal',
  is?:
    | 'div'
    | 'p'
    | 'span',
  size?:
    | '2xl'
    | 'md'
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
      size === 'md' && 'text-md',
      size === '2xl' && 'text-2xl',
      variant === 'secondary' && 'text-gray-600 dark:text-gray-400',
      fontWeight === 'bold' && 'font-bold',
    ]"
  >
    <slot />
  </component>
</template>
