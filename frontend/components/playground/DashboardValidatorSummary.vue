<script setup lang="ts">
</script>
<template>
  <div class="list">
    <div class="col">
      <DashboardTableValidators
        :validators="[123, 134, 5453, 34567, 12312, 12312, 6546, 543, 34567, 345]"
        context="group"
        :group-id="12"
      />
    </div>
    <div class="col">
      <DashboardTableValidators :validators="[123]" context="propsoal" :group-id="12" />
    </div>
    <div class="col">
      <DashboardTableValidators :validators="[123, 23]" context="sync" :group-id="12" />
    </div>
    <div class="col">
      <DashboardTableValidators :validators="[]" context="group" :group-id="12" />
    </div>
    <div class="col">
      <DashboardTableEfficiency :failed="10" :success="100" />
    </div>
    <div class="col">
      <DashboardTableEfficiency :failed="0" :success="100" />
    </div>
    <div class="col">
      <DashboardTableEfficiency :failed="30" :success="100" />
    </div>
    <div class="col">
      <DashboardTableEfficiency :failed="30" :success="0" />
    </div>
    <div class="col">
      <DashboardTableEfficiency :failed="0" :success="0" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.list {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;

  .col {
    max-width: 150px;
  }
}
</style>
