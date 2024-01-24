<script setup lang="ts"></script>
<template>
  <div class="container">
    {{ $t("test") }}
  </div>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 50px;
  height: 101px;
  border: 1px solid var(--container-border-color);
  border-radius: var(--border-radius);
  color: var(--container-color);
  background: var(--container-background);
  padding-left: var(--padding-xl);
  padding-right: var(--padding-xl);
}

.content {
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin);
}
</style>
