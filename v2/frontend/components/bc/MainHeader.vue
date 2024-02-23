<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLatestStateStore } from '~/stores/useLatestStateStore'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latest } = storeToRefs(useLatestStateStore())

</script>
<template>
  <div class="header top">
    <div class="content">
      <div>Current Epoch: {{ latest?.currentEpoch }}</div>
      <BcMainSearchBar v-if="!props.isHomePage" location="header" />
      <NuxtLink to="/login">
        Login
      </NuxtLink>
    </div>
  </div>
  <div class="header bottom">
    <div class="content">
      <NuxtLink to="/" class="">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </NuxtLink>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--light-grey);

  &.top {
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }

  &.bottom {
    height: var(--navbar2-height);
    background-color: var(--container-background);
    color: var(--container-color);
    border-bottom: 1px solid var(--container-border-color);
  }

  .content {
    width: var(--content-width);
    margin-left: var(--content-margin);
    margin-right: var(--content-margin);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}

.page {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>
