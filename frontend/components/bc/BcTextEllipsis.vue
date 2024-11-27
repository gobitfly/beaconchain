<script setup lang="ts"></script>

<template>
  <span class="bc-text-ellipsis">
    <slot />
  </span>
</template>

<style scoped lang="scss">
.bc-text-ellipsis {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
