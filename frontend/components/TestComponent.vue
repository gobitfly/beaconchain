<script setup lang="ts">
const { test } = defineProps<{
  test: any,
}>()
console.log('TestComponent rendered with props:', test)
</script>

<template>
  <div>
    Component: testComponent
    <pre>{{ test }}</pre>
  </div>
</template>

<style scoped></style>
