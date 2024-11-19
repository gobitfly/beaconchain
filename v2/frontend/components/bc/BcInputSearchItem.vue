<script setup lang="ts">
defineProps<{
  isAriaSelected: boolean,
}>()
</script>

<template>
  <li
    :aria-selected="isAriaSelected"
    role="option"
    class="bc-input-search__list-item"
    tabindex="-1"
  >
    <slot />
  </li>
</template>

<style scoped lang="scss">
.bc-input-search__list-item {
  grid-column: 1 / -1;
  display: grid;
  align-items: center;
  grid-template-columns: subgrid;
  padding: var(--padding);
  width: 100%;
  cursor: pointer;
  &:hover,
  &[aria-selected="true"] {
    background: var(--list-hover-background);
  }
}
</style>
