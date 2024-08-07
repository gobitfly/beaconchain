<script setup lang="ts">
</script>

<template>
  <NuxtLink :no-prefetch="true">
    <template #default>
      <slot />
    </template>
  </NuxtLink>
</template>
