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
    height: 24px;
    color: var(--text-color-disabled);
    background-color: var(--container-background);
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    width: 100%;
    padding-right: 21px;
  }

  .copy-icon {
    position: absolute;
    right: 3px;
    top: 3px;
  }
}
</style>
