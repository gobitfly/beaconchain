<script setup lang="ts">
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  is = 'button',
  size = 'md',
} = defineProps<{
  is?: 'a' | 'button',
  name: IconName,
  screenreaderText: TranslationInput,
  size?: 'md',
  variant: 'secondary',
}>()
</script>

<template>
  <component
    :is
    class="flex rounded-full bg-linear-to-b disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'secondary' && 'from-gray-300 to-gray-200 dark:from-charcoal-600 dark:to-charcoal-700 opacity-90',
      size === 'md' && 'p-md',
    ]"
  >
    <BaseScreenreaderOnly :screenreader-text />
    <BaseIcon
      :name
      class="w-6 h-6"
    />
  </component>
</template>

<style scoped></style>
