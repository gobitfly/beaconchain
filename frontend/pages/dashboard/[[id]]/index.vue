<script setup lang="ts">
import { warn } from 'vue'

const route = useRoute()

warn(`route id: ${route.params.id}, query: ${route.query}`)
</script>

<template>
  <BcPageWrapper>
    <template #top>
      <div class="h1 dashboard_title">
        {{ $t('dashboard.title') }}
      </div>
      <DashboardValidatorOverview class="overview" />
    </template>
    <div>
      <DashboardValidatorSlotViz />
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>

.dashboard_title, .overview{
  margin-bottom: var(--padding-large);
}

</style>
