<script setup lang="ts">
const { variant = 'primary' } = defineProps<{
  // isDisabled: boolean,
  variant?: 'primary' | 'secondary',
}>()
</script>

<template>
  <button
    class="bg-linear-to-b py-md px-sm rounded-full text-sm opacity-90 hover:opacity-95 disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80"
    :class="[
      variant === 'primary' && 'from-gray-100 to-gray-300 text-black',
    ]"
  >
    <span class="px-lg">
      <slot />
    </span>
  </button>
</template>
