<script setup lang="ts">
export type BaseHeadings = 'h2' | 'h3' | 'h4'
defineProps<{
  is: BaseHeadings,
  size:
    | 'lg'
    | 'md'
    | 'xs',
}>()
</script>

<template>
  <component
    :is
    class="font-bold text-balance"
    :class="[
      size === 'lg' && 'text-4xl tracking-[-.05rem]',
      size === 'md' && 'text-md',
      size === 'xs' && 'text-xl tracking-[-0.0063rem]',
    ]"
  >
    <slot />
  </component>
</template>

<style scoped></style>
