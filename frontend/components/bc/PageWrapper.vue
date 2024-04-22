<script setup lang="ts">
const props = defineProps({ isHomePage: { type: Boolean } })

</script>

<template>
  <div class="page">
    <BcHeaderMainHeader :is-home-page="props.isHomePage" />
    <div class="content">
      <slot name="top" />
      <BcAdControl />
      <slot />
      <div id="bottom_banner" />
      <slot name="bottom" />
    </div>
    <BcFooterMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.page {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.content {
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin);
}
</style>
