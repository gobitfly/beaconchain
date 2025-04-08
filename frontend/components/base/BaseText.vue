<script setup lang="ts">
withDefaults(defineProps<{
  tag?: 'div' | 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' | 'p' | 'span',
  variant?: '2xl' | 'base',
}>(), {
  tag: 'span',
  variant: 'base',
})
</script>

<template>
  <Component
    :is="tag"
    :class="{
      'mb-5 font-montserrat text-2xl font-medium dark:text-gray-50': variant === '2xl',
    }"
  >
    <slot />
  </Component>
</template>

<style scoped></style>
