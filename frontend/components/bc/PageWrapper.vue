<script setup lang="ts">
</script>
<template>
  <div class="page">
    <BcMainHeader />
    <div class="content">
      <slot />
    </div>
    <BcMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.page{
  display: flex;
  flex-direction: column;
  align-items: center;
}
.content{
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin) ;
}
</style>
