<script setup lang="ts">
import type { VDBConsolidationsClTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isMobile?: boolean,
  rejectReason?: VDBConsolidationsClTableRow['reject_reason'],
  status: VDBConsolidationsClTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-cl-consolidations-status">
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
      :class="{ 'width-overwrite-table': !isMobile }"
      color="red"
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
    <BcTooltip
      v-if="status === 'rejected' && !isMobile"
      tooltip-width="195px"
      tooltip-text-align="left"
      class="status-tooltip-trigger"
    >
      <BcIcon
        name="circle-info"
      />
      <template #tooltip>
        <slot name="tooltip">
          <span v-if="rejectReason === 'full_queue'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.full_queue') }}
          </span>
          <span v-if="rejectReason === 'insufficient_consolidation_churn'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.insufficient_consolidation_churn') }}
          </span>
          <span v-if="rejectReason === 'source_address_mismatch'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_address_mismatch') }}
          </span>
          <span v-if="rejectReason === 'source_equals_target'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_equals_target') }}
          </span>
          <span v-if="rejectReason === 'source_exiting'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_exiting') }}
          </span>
          <span v-if="rejectReason === 'source_inactive'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_inactive') }}
          </span>
          <span v-if="rejectReason === 'source_no_execution_withdrawal_credentials'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_no_execution_withdrawal_credentials') }}
          </span>
          <span v-if="rejectReason === 'source_pending_withdrawals'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_pending_withdrawals') }}
          </span>
          <span v-if="rejectReason === 'source_slashed'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_slashed') }}
          </span>
          <span v-if="rejectReason === 'source_too_young'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_too_young') }}
          </span>
          <span v-if="rejectReason === 'source_unknown_pubkey'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.source_unknown_pubkey') }}
          </span>
          <span v-if="rejectReason === 'target_exiting'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.target_exiting') }}
          </span>
          <span v-if="rejectReason === 'target_inactive'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.target_inactive') }}
          </span>
          <span v-if="rejectReason === 'target_not_compounding'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.target_not_compounding') }}
          </span>
          <span v-if="rejectReason === 'target_unknown_pubkey'">
            {{ $t('dashboard.validator.cl_consolidations.reject_reason.target_unknown_pubkey') }}
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
.dashboard-table-cl-consolidations-status {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}
</style>
