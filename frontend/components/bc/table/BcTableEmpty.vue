<template>
  <div
    class="bc-table-empty delayed-fadein-animation"
  >
    <span class="big_text">
      <slot />
    </span>
    <slot name="icon" />
  </div>
</template>

<style lang="scss" scoped>
.bc-table-empty {
  display: flex;
  width: 100%;
  height: 400px;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: var(--text-color-disabled);
  gap: var(--padding);
  cursor: pointer;
}
</style>
