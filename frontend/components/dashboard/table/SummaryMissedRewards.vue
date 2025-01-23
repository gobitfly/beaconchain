<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { VDBGroupSummaryMissedRewards } from '~/types/api/validator_dashboard'

interface Props {
  missedRewards: VDBGroupSummaryMissedRewards,
}
const props = defineProps<Props>()
const {
  formatAmount,
} = useCurrency()
const currencyItems = computed(() =>
  [
    {
      consensusLayerValue: props.missedRewards.proposer_rewards.cl,
    },
    {
      executionLayerValue: props.missedRewards.proposer_rewards.el,
    },
    {
      consensusLayerValue: props.missedRewards.attestations,
    },
    {
      executionLayerValue: props.missedRewards.sync,
    },
  ])
</script>

<template>
  <div
    class="summary-reward"
  >
    <BcFormatAmount
      :currency-items
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
              :currency-items="[
                {
                  executionLayerValue: missedRewards.proposer_rewards.el,
                },
              ]"
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.cl_rewards") }}:
            </span>
            <BcFormatAmount
              :currency-items="[
                {
                  consensusLayerValue: missedRewards.proposer_rewards.cl,
                },
              ]"
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.attestations") }}:
            </span>
            {{ formatAmount(missedRewards.attestations, {
              hasHigherPrecision: true,
            }) }}
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.summary.row.sync_committee") }}:
            </span>
            <BcFormatAmount
              :currency-items="[
                {
                  consensusLayerValue: missedRewards.sync,
                },
              ]"
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
