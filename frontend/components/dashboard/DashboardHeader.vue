<script lang="ts" setup>
const emit = defineEmits<{(e: 'showCreation'): void }>()

</script>

<template>
  <div class="header-container">
    <div class="h1 dashboard-title">
      {{ $t('dashboard.title') }}
    </div>
    <div class="dashboard-buttons">
      <DashboardSelectionButton />
      <DashboardSelectionButton />
      <DashboardSelectionButton />
      <Button class="p-button-icon-only" @click="emit('showCreation')">
        <IconPlus alt="Plus icon" width="100%" height="100%" />
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header-container {
  display: flex;
  justify-content: space-between;

  .dashboard-title {
    margin-bottom: var(--padding-large);
  }

  .dashboard-buttons {
    display: flex;
    gap: var(--padding);
  }
}
</style>
