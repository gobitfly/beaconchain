<script setup lang="ts">
defineProps<{
  ariaSelected: boolean,
}>()
</script>

<template>
  <li
    :aria-selected
    role="option"
    class="bc-input-search__list-item"
    tabindex="-1"
  >
    <slot />
  </li>
</template>

<style scoped></style>
