<script setup lang="ts">
import { useLatestState } from '~/stores/useLatestState'
const { getLatestState } = useLatestState()
await useAsyncData('latest_state', () => getLatestState())

</script>
<template>
  <BcPageWrapper>
    <NuxtLink to="/login" class="test-link">
      Login
    </NuxtLink>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.test-link{
  background-color: var(--text-color);
  color: var(--background-color);
}
</style>
