<script setup lang="ts">
const year = new Date().getFullYear()
const version = '2'
</script>

<template>
  <span>© bitfly GmbH {{ year }} | v{{ version }}</span>
</template>
