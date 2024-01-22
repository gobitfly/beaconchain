<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLatestState } from '~/stores/useLatestState'

const { latest } = storeToRefs(useLatestState())

</script>
<template>
  <div>Current Epoch: {{ latest?.currentEpoch }}</div>
</template>

<style lang="scss" scoped>
</style>
