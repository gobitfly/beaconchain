<script setup lang="ts">
interface Props {
  label?: string, // will default to value
  value?: string,
}
defineProps<Props>()
</script>

<template>
  <div
    v-if="value"
    class="copy-label"
  >
    <input
      :disabled="true"
      :value="label || value"
    >
    <BcCopyToClipboard
      :value
      class="copy-icon"
    />
  </div>
</template>

<style>
.copy-label {
  position: relative;

  input {
    color: var(--text-color-disabled);
    background-color: var(--container-background);
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    width: 100%;
    padding: var(--padding);
    padding-right: 2rem;
  }

  .copy-icon {
    position: absolute;
    right: var(--padding);
    bottom: var(--padding);
  }
}
</style>
