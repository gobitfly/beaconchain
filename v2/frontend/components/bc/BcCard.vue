<script setup lang="ts"></script>

<template>
  <div class="bc-card">
    <slot />
    <slot name="floating-action-button" />
  </div>
</template>

<style scoped lang="scss">
.bc-card {
  background-color: var(--input-background);
  border-radius: var(--corner-radius, 4px);
  padding: 0.625rem;
  border: 1px solid var(--input-border-color);
}
</style>
