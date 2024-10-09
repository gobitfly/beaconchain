<script setup lang="ts">
</script>

<template>
  <div class="bc-settings">
    <slot />
    <BcInputCheckbox />
  </div>
</template>

<style scoped lang="scss">
.bc-settings{
  display: grid;
  column-gap: 0.625rem;
  grid-template-columns: 1fr auto min-content;
}
</style>
