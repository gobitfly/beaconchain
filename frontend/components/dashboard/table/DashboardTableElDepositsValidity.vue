<script setup lang="ts">
import type { VDBExecutionDepositsTableRow } from '~/types/api/validator_dashboard'

defineProps<{
  isMobile?: boolean,
  validity: VDBExecutionDepositsTableRow['validity'],
}>()
</script>

<template>
  <div class="dashboard-table-el-deposits-validity">
    <BcBadge
      v-if="validity === 'valid'"
      :class="{ 'width-overwrite-table': !isMobile }"
      color="green"
    >
      <template v-if="isMobile">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.valid" />
        <BcIcon
          name="check"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.valid") }}
      </template>
    </BcBadge>
    <BcBadge
      v-if="validity === 'invalid_skipped'"
      color="green"
      :class="{ 'width-overwrite-table': !isMobile }"
    >
      <template v-if="isMobile">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.valid" />
        <BcIcon
          name="check"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.valid") }}
      </template>
    </BcBadge>
    <BcTooltip
      v-if="validity === 'invalid_skipped' && !isMobile"
      tooltip-width="195px"
      tooltip-text-align="left"
      class="status-tooltip-trigger"
    >
      <BcIcon
        name="circle-info"
      />
      <template #tooltip>
        <slot name="tooltip">
          {{ $t('dashboard.validator.el_deposits.info_valid_with_invalid_signature') }}
        </slot>
      </template>
    </BcTooltip>
    <BcBadge
      v-if="validity === 'invalid'"
      color="red"
      :class="{ 'width-overwrite-table': !isMobile }"
    >
      <template v-if="isMobile">
        <BcScreenreaderOnly screenreader-text="dashboard.validator.table.status_text.invalid" />
        <BcIcon
          name="xmark"
          aria-hidden="true"
        />
      </template>
      <template v-else>
        {{ $t("dashboard.validator.table.status_text.invalid") }}
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
.dashboard-table-el-deposits-validity {
  display: flex;
  gap: var(--padding-medium);
  align-items: center;
}
</style>
