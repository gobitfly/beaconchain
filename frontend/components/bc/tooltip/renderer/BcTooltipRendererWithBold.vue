<script setup lang="ts">
defineProps<{
  texts?: string[], // a list of texts where every second item will be bold, empty strings will not be added
}>()

</script>

<template>
  <template v-for="(text, index) in texts" :key="index">
    <span v-if="text" :class="{bold: index%2}">{{ text }}</span>
  </template>
</template>
