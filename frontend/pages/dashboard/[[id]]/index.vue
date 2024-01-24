<script setup lang="ts">

const route = useRoute()

console.log(`route id: ${route.params.id}, query: ${route.query}`)

</script>

<template>
  <BcPageWrapper>
    <div>
      <div class="h1">
        {{ $t('dashboard.title') }}
        <DashboardValidatorOverview />
      </div>
    </div>
  </BcPageWrapper>
</template>
