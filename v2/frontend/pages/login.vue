<script setup lang="ts">
const test = ref('blah')
</script>

<template>
  <div>{{ $t("login.title") }} {{ test }}</div>
</template>
