<script setup lang="ts">
defineProps<{
  layout: '1fr_1fr_1fr' | '2fr_1fr',
}>()
</script>

<template>
  <div
    class="grid gap-xl"
    :class="[
      layout === '2fr_1fr' && 'md:grid-cols-[2fr_1fr] md:[&>*:nth-child(3)]:col-span-2',
      layout === '1fr_1fr_1fr' && 'md:grid-cols-[1fr_1fr_1fr]',
    ]"
  >
    <slot />
  </div>
</template>

<style scoped></style>
