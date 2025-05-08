<script setup lang="ts">
import type { VDBConsensusDepositsTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isCompact?: boolean,
  rejectReason?: VDBConsensusDepositsTableRow['reject_reason'],
  status: VDBConsensusDepositsTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-cl-deposits-status">
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
      v-if="status === 'completed'"
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
        {{ $t("dashboard.validator.table.status_text.completed") }}
      </template>
    </BcBadge>
    <BcBadge
      v-if="status === 'rejected'"
      color="red"
      :class="{ 'width-overwrite-table': !isCompact }"
    >
      <template v-if="isCompact">
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
    <BcTooltip
      v-if="rejectReason === 'invalid_signature' && status === 'rejected' && !isCompact"
      tooltip-width="195px"
      tooltip-text-align="left"
      class="status-tooltip-trigger"
    >
      <BcIcon
        name="circle-info"
      />
      <template #tooltip>
        <slot name="tooltip">
          <span v-if="rejectReason === 'invalid_signature'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.invalid_signature') }}
          </span>
        </slot>
      </template>
    </BcTooltip>
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
