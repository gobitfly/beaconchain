<script setup lang="ts">
import { useLatestState } from '~/stores/useLatestState'

const { latest } = useLatestState()

</script>
<template>
  <div>Current Epoch: {{ latest?.currentEpoch }}</div>
</template>

<style lang="scss" scoped>
</style>
