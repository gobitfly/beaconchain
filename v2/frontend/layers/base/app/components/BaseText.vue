<script setup lang="ts">
const {
  is = 'span',
  size = 'md',
} = defineProps<{
  dimmed?: boolean,
  is?:
    | 'span',
  size?:
    | '2xl'
    | 'sm',
}>()
</script>

<template>
  <component
    :is
    class="font-semibold text-pretty"
    :class="[
      size === 'sm' && 'text-sm',
      size === '2xl' && 'text-2xl',
      dimmed && 'text-gray-400',
    ]"
  >
    <slot />
  </component>
</template>
