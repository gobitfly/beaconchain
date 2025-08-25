<script setup lang="ts"></script>

<template>
  <div class="page">
    <header>
      <slot name="header" />
    </header>
    <slot name="banner" />
    <main class="content">
      <slot />
    </main>
    <footer>
      <TheFooter />
    </footer>
  </div>
</template>

<style lang="scss" scoped>
.page {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100vh;
}

.content {
  max-width: var(--content-width);
  margin: var(--padding) auto
}
</style>
