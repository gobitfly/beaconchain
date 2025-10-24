<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'dark:bg-black dark:text-white bg-white text-black transition-colors duration-1500 ',
  },
  htmlAttrs: { class: 'base-layout' },
})
</script>

<template>
  <div class="grid h-dvh grid-rows-[auto_1fr_auto] font-urbanist antialiased">
    <header>
      <slot name="header" />
    </header>
    <main>
      <slot />
    </main>
    <footer class="mt-11xl">
      <TheFooter />
    </footer>
  </div>
</template>
