<script setup lang="ts">
const { gap = 'xl' } = defineProps<{
  gap?: '5xl' | 'xl',
  layout: '1fr_1fr_1fr' | '2fr_1fr',
}>()
</script>

<template>
  <div
    class="grid"
    :class="[
      layout === '2fr_1fr' && 'md:grid-cols-[2fr_1fr] md:[&>*:nth-child(3)]:col-span-2',
      layout === '1fr_1fr_1fr' && 'md:grid-cols-[1fr_1fr_1fr]',
      gap === 'xl' && 'gap-xl',
      gap === '5xl' && 'gap-5xl',
    ]"
  >
    <slot />
  </div>
</template>

<style scoped></style>
