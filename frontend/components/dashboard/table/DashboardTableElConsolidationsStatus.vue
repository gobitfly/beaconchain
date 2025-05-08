<script setup lang="ts">
import type { VDBConsolidationsElTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isCompact?: boolean,
  status: VDBConsolidationsElTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-el-consolidations-status">
    <BcBadge
      v-if="status === 'queued'"
      color="orange"
      :class="{ 'width-overwrite-table': !isCompact }"
    >
      <template v-if="isCompact">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.queued" />
        <BcIcon
          name="sync"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.queued") }}
      </template>
    </BcBadge>
    <BcBadge
      v-if="status === 'processed'"
      :class="{ 'width-overwrite-table': !isCompact }"
      color="green"
    >
      <template v-if="isCompact">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.completed" />
        <BcIcon
          name="check"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.processed") }}
      </template>
    </BcBadge>
  </div>
</template>

<style scoped lang="scss">
.width-overwrite-table {
  width: 5rem;
}
.status-tooltip-trigger {
  display: flex;
}
.dashboard-table-el-consolidations-status {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}
</style>
