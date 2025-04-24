<script setup lang="ts">
import type { VDBWithdrawalsElTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isMobile?: boolean,
  status: VDBWithdrawalsElTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-el-withdrawals-status">
    <BcBadge
      v-if="status === 'queued'"
      color="orange"
      :class="{ 'width-overwrite-table': !isMobile }"
    >
      <template v-if="isMobile">
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
      :class="{ 'width-overwrite-table': !isMobile }"
      color="green"
    >
      <template v-if="isMobile">
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
.dashboard-table-el-withdrawals-status {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}
</style>
