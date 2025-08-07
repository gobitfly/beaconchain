<script setup lang="ts">
defineProps<{
  description: TranslationKey,
  // statusCode: number,
  title: TranslationKey,
}>()
const error = useError()
</script>

<template>
  <div class="bc-error">
    <h1>{{ $t(title) }} ({{ error?.statusCode }})</h1>
    <p>{{ $t(description) }}</p>
    <slot />
  </div>
</template>

<style scoped>
.bc-error {
  display: flex;
  gap: 1rem;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
</style>
