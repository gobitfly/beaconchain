<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'dark:bg-black dark:text-white bg-white text-black',
  },
  htmlAttrs: { class: 'base-layout' },
})
</script>

<template>
  <div class="font-urbanist antialiased grid h-dvh grid-rows-[auto_1fr_auto]">
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
