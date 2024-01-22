<script setup lang="ts">
import { useLatestState } from '~/stores/useLatestState'
const latestStateStore = useLatestState()
const { getLatestState } = latestStateStore
await useAsyncData('latest_state', () => getLatestState())

</script>
<template>
  <BcPageWrapper>
    <NuxtLink to="/login" class="test-link">
      Login
    </NuxtLink>
    <Button label="test" @click="()=>getLatestState()" />
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.test-link{
  background-color: var(--text-color);
  color: var(--background-color);
}
</style>
