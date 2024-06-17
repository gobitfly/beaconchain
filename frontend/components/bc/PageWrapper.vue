<script setup lang="ts">
const props = defineProps({ isHomePage: { type: Boolean } })
</script>

<template>
  <BcCookieModal />
  <div class="page">
    <BcHeaderMainHeader :is-home-page="props.isHomePage" />
    <BcMaintenanceBanner />
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
  position: relative;
  flex-direction: column;
  align-items: center;
  overflow-x: hidden;
}

.content {
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin);
}
</style>
