<script lang="ts" setup>
import type { DashboardType } from '~/types/dashboard'
import type { Dashboard } from '~/types/api/dashboard'

const { t: $t } = useI18n()

interface Props {
  type: DashboardType | 'notifications',
  dashboards?: Dashboard[]
}
const props = defineProps<Props>()

</script>

<template>
  <div v-if="props.type=='notifications'">
    {{ $t('dashboard.notifications') }}
  </div>
  <div v-else>
    <Button v-for="dashboard in dashboards" :key="dashboard.id" class="p-button-text" :label="dashboard.name" />
  </div>
</template>

<style lang="scss">

</style>
