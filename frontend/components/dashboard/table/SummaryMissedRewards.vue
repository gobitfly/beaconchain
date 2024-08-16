<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { VDBGroupSummaryMissedRewards } from '~/types/api/validator_dashboard'

interface Props {
  missedRewards?: VDBGroupSummaryMissedRewards,
}
const props = defineProps<Props>()

const total = computed(() =>
  props.missedRewards
    ? convertSum(
      props.missedRewards.proposer_rewards.cl,
      props.missedRewards.proposer_rewards.el,
      props.missedRewards.attestations,
      props.missedRewards.sync,
    )
    : undefined,
)
</script>

<template>
  <div
    v-if="total && !total.isZero()"
    class="summary-reward"
  >
    <BcFormatValue
      :value="total"
      :no-tooltip="true"
    />
    <BcTooltip :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div>
          <div class="tt-row">
            <span class="bold top">{{ $t("dashboard.validator.summary.tooltip.estimated_loss") }}
            </span>
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.el_rewards") }}:
            </span>
            <BcFormatValue
              :value="missedRewards?.proposer_rewards.el"
              :no-tooltip="true"
              :full-value="true"
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.cl_rewards") }}:
            </span>
            <BcFormatValue
              :value="missedRewards?.proposer_rewards.cl"
              :no-tooltip="true"
              :full-value="true"
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.attestations") }}:
            </span>
            <BcFormatValue
              :value="missedRewards?.attestations"
              :no-tooltip="true"
              :full-value="true"
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.sync_committee") }}:
            </span>
            <BcFormatValue
              :value="missedRewards?.sync"
              :no-tooltip="true"
              :full-value="true"
            />
          </div>
        </div>
      </template>
    </BcTooltip>
  </div>
  <div v-else>
    -
  </div>
</template>

<style lang="scss" scoped>
.tt-row {
  display: flex;
  flex-wrap: nowrap;
  white-space: nowrap;
  gap: 3px;
  .top {
    padding-bottom: var(--padding);
  }
}

.summary-reward {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--padding-small);
}
</style>
