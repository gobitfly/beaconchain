<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ClElValue } from '~/types/api/common'

const { reward } = defineProps<{
  reward: ClElValue<string>,
}>()
const hasReward = computed(() => !(reward.el === '0' && reward.cl === '0'))
</script>

<template>
  <div
    class="summary-reward"
  >
    <BcFormatAmount
      :currency-items="[{
        executionLayerValue: reward.el,
        consensusLayerValue: reward.cl,
      }]"
      has-color
    />
    <BcTooltip
      v-if="hasReward"
      :fit-content="true"
    >
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.el_rewards") }}:</span>
            <BcFormatAmount
              :currency-items="[{
                executionLayerValue: reward.el,
              }]"
              has-higher-precision
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.cl_rewards") }}:</span>
            <BcFormatAmount
              :currency-items="[{
                consensusLayerValue: reward.cl,
              }]"
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
}

.summary-reward {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--padding-small);
}
</style>
