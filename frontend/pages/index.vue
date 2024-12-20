<script setup lang="ts">
useBcSeo()
</script>

<template>
  <BcPageWrapper :is-home-page="true">
    <h1>Index page</h1>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.search-bar {
  display: flex;
  justify-content: center;
  .bar {
    width: 730px;
  }
}
</style>
