<script setup lang="ts">
import { useDashboardGroupsStore } from '~/stores/dashboard/useDashboardGroupsStore'

const store = useDashboardGroupsStore()
const { getGroups } = store

const { groups } = storeToRefs(store)
await useAsyncData('validator_dashboard_overview', () => getGroups())

</script>
<template>
  {{ groups }}
</template>

<style lang="scss" scoped>
</style>
