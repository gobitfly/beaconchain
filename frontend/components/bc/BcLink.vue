<script setup lang="ts">
</script>
<template>
  <NuxtLink class="link" :no-prefetch="true">
    <template #default>
      <slot />
    </template>
  </NuxtLink>
</template>
