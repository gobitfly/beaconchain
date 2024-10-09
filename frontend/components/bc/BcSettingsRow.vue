<script setup lang="ts">
</script>

<template>
  <div class="bc-settings-row">
    <slot />
  </div>
</template>

<style scoped lang="scss">
.bc-settings-row {
  display: grid;
  grid-template-columns: subgrid;
}
</style>
