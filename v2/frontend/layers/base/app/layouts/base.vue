<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'dark:bg-black bg-white',
  },
})
</script>

<template>
  <div class="font-urbanist antialiased grid h-dvh grid-rows-[auto_1fr_auto]">
    <header>
      <BaseNavigation />
    </header>
    <main>
      <slot />
    </main>
    <TheFooter>
      Here goes the footer
    </TheFooter>
  </div>
</template>
