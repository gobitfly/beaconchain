<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { VDBGroupSummaryMissedRewards } from '~/types/api/validator_dashboard'

const { missedRewards } = defineProps<{
  missedRewards: VDBGroupSummaryMissedRewards,
}>()
</script>

<template>
  <div
    class="summary-reward"
  >
    <BcFormatAmount
      :currency-items=" [
        {
          consensusLayerValue: missedRewards.proposer_rewards.cl,
        },
        {
          executionLayerValue: missedRewards.proposer_rewards.el,
        },
        {
          consensusLayerValue: missedRewards.attestations,
        },
        {
          consensusLayerValue: missedRewards.sync,
        },
      ]"
      target-unit-crypto="auto"
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
            <BcFormatAmount
              :value="missedRewards.proposer_rewards.el"
              source-currency="elCurrency"
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.cl_rewards") }}:
            </span>
            <BcFormatAmount
              :value="missedRewards.proposer_rewards.cl"
              has-additional-selected-currency-main
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.attestations") }}:
            </span>
            <BcFormatAmount
              :value="missedRewards.attestations"
              has-additional-selected-currency-main
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.sync_committee") }}:
            </span>
            <BcFormatAmount
              :value="missedRewards.sync"
              has-additional-selected-currency-main
              has-higher-precision
            />
          </div>
        </div>
      </template>
    </BcTooltip>
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
