<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLatestState } from '~/stores/useLatestState'

const { latest } = storeToRefs(useLatestState())

</script>
<template>
  <div class="header top">
    <div>Current Epoch: {{ latest?.currentEpoch }}</div>

    <NuxtLink to="/login">
      Login
    </NuxtLink>
  </div>
  <div class="header bottom">
    <NuxtLink to="/" class="">
      <IconBeaconchainLogo alt="Beaconcha.in logo" />
    </NuxtLink>
  </div>
</template>

<style lang="scss" scoped>
.header{
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--light-grey);

  &.top{
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }
  &.bottom{
    height: var(--navbar2-height);
    background-color: var(--container-background);
    color: var(--container-color);
    border-bottom: 1px solid var(--container-border-color);
  }
}

</style>
