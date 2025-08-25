<script setup lang="ts">
</script>

<template>
  <div
    class="minimalist"
  >
    <div class="top-background" />
    <div class="rows">
      <BcNavigationLogo layout-adaptability="low" />
    </div>
  </div>
</template>

<style scoped lang="scss">
$mobileHeaderThreshold: 600px;

.minimalist {
  color: var(--header-top-font-color);
  position: relative;
  display: flex;
  width: 100%;
  justify-content: center;
  .top-background {
    position: absolute;
    width: 100%;
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }
  .rows {
    width: var(--content-width);
  }
  @media (max-width: $mobileHeaderThreshold) {
    .top-background {
      height: 36px;
    }
  }
}
</style>
