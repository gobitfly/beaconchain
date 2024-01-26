<script setup lang="ts">
</script>
<template>
  <div class="page">
    <BcMainHeader />
    <div class="content">
      <slot name="top" />
      <div class="banner">
        Banner top Placeholder (to be replaced with dynamic banner component)
      </div>
      <slot />
      <div class="banner">
        Banner bottom Placeholder (to be replaced with dynamic banner component)
      </div>
      <slot name="bottom" />
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

//TODO: remove once we have a 'real' banner component
.banner{
  height: 90px;
  background-color: #66BCE9;
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: var(--content-margin);
}
</style>
