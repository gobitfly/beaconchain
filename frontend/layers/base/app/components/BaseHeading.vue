<script setup lang="ts">
defineProps<{
  is:
    | 'h2',
  size: 'lg',
}>()
</script>

<template>
  <component
    :is
    class="font-semibold text-balance"
    :class="[
      size === 'lg' && 'text-4xl tracking-[-.05rem]',
    ]"
  >
    <slot />
  </component>
</template>

<style scoped></style>
