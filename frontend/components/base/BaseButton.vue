<script setup lang="ts">
type ButtonVariant = 'disabled' | 'primary' | 'secondary'
const props = withDefaults(
  defineProps<{
    isDisabled?: boolean,
    variant?: ButtonVariant,
  }>(),
  {
    isDisabled: false,
    variant: 'primary',
  },
)

const buttonVariant = computed<ButtonVariant>(() => {
  if (props.isDisabled) return 'disabled'
  if (props.variant === 'secondary') return 'secondary'
  return 'primary'
})
</script>

<template>
  <button
    :class="{
      'bg-orange-400 hover:bg-orange-500': buttonVariant === 'primary',
      'bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:font-normal dark:text-gray-100 dark:hover:bg-gray-800':
        buttonVariant === 'secondary',
      'cursor-not-allowed bg-gray-400 text-gray-100': buttonVariant === 'disabled',
    }"
    class="w-36 truncate rounded px-2 py-1 text-sm font-medium leading-5 text-gray-800"
  >
    <slot />
  </button>
</template>

<style scoped></style>
