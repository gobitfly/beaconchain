<script setup lang="ts">
defineProps<{ hasBorderTop?: boolean }>()
</script>

<template>
  <div
    class="bc-list-section"
    :class="{ 'has-border-top': hasBorderTop }"
  >
    <slot />
  </div>
</template>

<style scoped lang="scss">
$gap: 15px;
.bc-list-section {
  display: grid;
  padding-block: $gap;
  grid-template-columns: 3fr auto 2fr;
  gap: $gap;
  align-items: center;
}
.has-border-top {
  border-top: 1px solid var(--dark-grey);
}
</style>
