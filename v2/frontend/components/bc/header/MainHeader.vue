<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latestState, refreshLatestState } = useLatestStateStore()
await useAsyncData('latest_state', () => refreshLatestState())

</script>
<template>
  <div class="header top">
    <div class="content">
      <div>Current Epoch: {{ latestState?.currentEpoch }}</div>
      <BcSearchbarGeneral v-if="!props.isHomePage" bar-style="discreet" />
      <NuxtLink to="/login">
        Login
      </NuxtLink>
    </div>
  </div>
  <div class="header bottom">
    <div class="content">
      <NuxtLink to="/" class="logo">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </NuxtLink>

      <BcHeaderMegaMenu />
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
    min-height: var(--navbar2-height);
    background-color: var(--container-background);
    color: var(--container-color);
    border-bottom: 1px solid var(--container-border-color);
  }

  .content {
    width: var(--content-width);
    margin-left: var(--content-margin);
    margin-right: var(--content-margin);
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }

  .logo {
    height: var(--navbar2-height);
    display: flex;
    align-items: center;
  }
}

.page {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>
