<script setup lang="ts">
const { variant = 'primary' } = defineProps<{
  variant?: 'primary' | 'secondary',
}>()
</script>

<template>
  <button
    class="bg-linear-to-b py-md px-sm rounded-full text-sm font-semibold opacity-90 hover:opacity-95 disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'primary' && 'from-gray-100 to-gray-300 text-black',
    ]"
  >
    <span class="px-lg">
      <slot />
    </span>
  </button>
</template>
