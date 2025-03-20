<script setup lang="ts">
const { is = 'span' } = defineProps<{
  is?:
    'div'
    | 'h1'
    | 'h2'
    | 'h3'
    | 'h4'
    | 'h5'
    | 'h6'
    | 'legend'
    | 'p'
    | 'span',
}>()
</script>

<template>
  <component
    :is
    class="bc-screenreader-only"
  >
    <slot />
  </component>
</template>

<style scoped lang="scss">
// see https://tailwindcss.com/docs/screen-readers
.bc-screenreader-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
</style>
