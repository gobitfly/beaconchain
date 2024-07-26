<script setup lang="ts">
const { count, keys } = useVarTest('accordeon')
</script>

<template>
  <div class="content">
    Accordeon: {{ count }}, keys: {{ keys.join(',') }}
    <BcFaq translation-path="faq.playground" />
  </div>
</template>

<style lang="scss" scoped>
.content {
  margin: 40px;
}
</style>
