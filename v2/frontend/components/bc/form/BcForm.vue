<script lang="ts" setup></script>

<template>
  <form class="bc-form">
    <slot />
  </form>
</template>

<style lang="scss" scoped>
.bc-form {
  display: flex;
  flex-direction: column;
  row-gap: var(--padding-small);
}
</style>
