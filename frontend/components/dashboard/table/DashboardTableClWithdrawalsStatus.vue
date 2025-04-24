<script setup lang="ts">
import type { VDBWithdrawalsClTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isMobile?: boolean,
  rejectReason?: VDBWithdrawalsClTableRow['reject_reason'],
  status: VDBWithdrawalsClTableRow['status'],
}>()
</script>

<template>
  <div class="dashboard-table-cl-withdrawals-status">
    <BcBadge
      v-if="status === 'queued'"
      color="orange"
      :class="{ 'dashboard-table-cl-withdrawals-status__badge': !isMobile }"
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
      :class="{ 'dashboard-table-cl-withdrawals-status__badge': !isMobile }"
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
      :class="{ 'dashboard-table-cl-withdrawals-status__badge': !isMobile }"
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
      tooltip-width="266px"
      tooltip-text-align="left"
      class="dashboard-table-cl-withdrawals-status__tooltip-trigger"
      tooltip-class="dashboard-table-cl-withdrawals-status__tooltip"
    >
      <BcIcon
        name="circle-info"
      />
      <template #tooltip>
        <slot name="tooltip">
          <span v-if="rejectReason === 'address_mismatch'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.address_mismatch') }}
          </span>
          <span v-if="rejectReason === 'excess_balance'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.excess_balance') }}
          </span>
          <span v-if="rejectReason === 'exiting'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.exiting') }}
          </span>
          <span v-if="rejectReason === 'full_queue'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.full_queue') }}
          </span>
          <span v-if="rejectReason === 'inactive'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.inactive') }}
          </span>
          <span v-if="rejectReason === 'insufficient_effective_balance'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.insufficient_effective_balance') }}
          </span>
          <span v-if="rejectReason === 'no_execution_withdrawal_credentials'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.no_execution_withdrawal_credentials') }}
          </span>
          <span v-if="rejectReason === 'not_compounding'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.not_compounding') }}
          </span>
          <span v-if="rejectReason === 'pending_withdrawals'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.pending_withdrawals') }}
          </span>
          <span v-if="rejectReason === 'too_young'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.too_young') }}
          </span>
          <span v-if="rejectReason === 'unknown_pubkey'">
            {{ $t('dashboard.validator.cl_withdrawals.reject_reason.unknown_pubkey') }}
          </span>
        </slot>
      </template>
    </BcTooltip>
  </div>
</template>

<style scoped lang="scss">
.dashboard-table-cl-withdrawals-status {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}

.dashboard-table-cl-withdrawals-status__badge {
  width: 5rem;
}

.dashboard-table-cl-withdrawals-status__tooltip-trigger {
  display: flex;
}

:global(.dashboard-table-cl-withdrawals-status__tooltip) {
  white-space: pre-line;
}
</style>
