<script setup lang="ts">
import type { VDBConsensusDepositsTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isMobile?: boolean,
  status: VDBConsensusDepositsTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-cl-deposits-status">
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
      v-if="status === 'completed'"
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
        {{ $t("dashboard.validator.table.status_text.completed") }}
      </template>
    </BcBadge>
    <BcBadge
      v-if="status === 'rejected'"
      color="red"
      :class="{ 'width-overwrite-table': !isMobile }"
    >
      <template v-if="isMobile">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.rejected" />
        <BcIcon
          name="xmark"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.rejected") }}
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
.dashboard-table-cl-deposits-status {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}
</style>
