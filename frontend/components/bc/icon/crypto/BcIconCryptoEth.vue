<script setup lang="ts">
const { colorMode = 'colored' } = defineProps<{
  colorMode?: 'colored' | 'currentColor' | 'dark' | 'light',
}>()

const fill = computed(() => {
  if (colorMode === 'currentColor') return 'currentColor'
  if (colorMode === 'dark') return '#111827'
  if (colorMode === 'light') return '#F3F4F6'
  return undefined
})
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 12 12"
  ><path
    :fill="fill ?? '#343434'"
    d="m6 0-.08.27v7.94l.08.08L9.5 6.1 6 0Z"
  /><path
    :fill="fill ?? '#8C8C8C'"
    d="M6 0 2.5 6.11 6 8.3V0Z"
  /><path
    :fill="fill ?? '#3C3C3B'"
    d="m6 8.99-.05.05v2.83L6 12l3.5-5.19L6 8.99Z"
  /><path
    :fill="fill ?? '#8C8C8C'"
    d="M6 12V8.99L2.5 6.8 6 12Z"
  /><path
    :fill="fill ?? '#141414'"
    d="M6 8.29 9.5 6.1 6 4.44v3.85Z"
  /><path
    :fill="fill ?? '#393939'"
    d="M2.5 6.11 6 8.3V4.44L2.5 6.1Z"
  /></svg>
</template>
