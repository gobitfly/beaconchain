<script setup lang="ts"></script>

<template>
  <div class="bc-tab-panel">
    <div class="bc-tab-panel__content">
      <slot />
    </div>
  </div>
</template>

<style scoped lang="scss">
.bc-tab-panel {
  height: 100%;
  padding-top: 30px;
  padding: var(--padding);

  &__content {
    max-width: 500px;
    margin-left: auto;
    margin-right: auto;
  }
}
</style>
