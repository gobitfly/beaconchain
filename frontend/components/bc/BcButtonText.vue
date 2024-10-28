<script setup lang="ts"></script>

<template>
  <button class="bc-button-text">
    <slot />
  </button>
</template>

<style scoped lang="scss">
  .bc-button-text {
    background: none;
    border: none;
    padding: 0;
    color: inherit;
    font-family: inherit;
    font-weight: inherit;
    font-size: inherit;
  }
</style>
