<script setup lang="ts">
defineProps<{
  is:
    | 'h2'
    | 'h3',
  size:
    | 'lg'
    | 'md',
}>()
</script>

<template>
  <component
    :is
    class="font-semibold text-balance"
    :class="[
      size === 'lg' && 'text-4xl tracking-[-.05rem]',
      size === 'md' && 'text-md',
    ]"
  >
    <slot />
  </component>
</template>

<style scoped></style>
