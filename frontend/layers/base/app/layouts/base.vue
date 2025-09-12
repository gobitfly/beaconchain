<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'dark:bg-black dark:text-white bg-white text-black',
  },
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
    <TheFooter>
      Here goes the footer
    </TheFooter>
  </div>
</template>
