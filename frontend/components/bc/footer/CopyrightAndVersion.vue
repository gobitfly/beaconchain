<script setup lang="ts">
const year = new Date().getFullYear()
const version = useRuntimeConfig().public.gitVersion
</script>

<template>
  <span>© bitfly GmbH {{ year }} | v{{ version }}</span>
</template>
