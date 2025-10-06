<script setup lang="ts">
export type BaseHeadings = 'h1' | 'h2' | 'h3' | 'h4'
defineProps<{
  is: BaseHeadings,
  size:
    | '2xl'
    | 'lg'
    | 'md'
    | 'xl'
    | 'xs'
    | 'xxs',
}>()
</script>

<template>
  <component
    :is
    class="font-bold text-balance"
    :class="[
      size === '2xl' && 'text-6xl tracking-[-.09rem]',
      size === 'xl' && 'text-5xl tracking-[-.075rem]',
      size === 'lg' && 'text-4xl tracking-[-.05rem]',
      size === 'xs' && 'text-xl tracking-[-0.0063rem]',
      size === 'xxs' && 'text-md',
    ]"
  >
    <slot />
  </component>
</template>

<style scoped></style>
