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
      target-unit-crypto="auto"
    />
    <BcTooltip
      v-if="hasReward"
      :fit-content="true"
    >
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div>
          <div>
            <h3 class="bold">
              {{ $t("dashboard.validator.blocks.el_rewards") }}
            </h3>
            <BcFormatAmount
              :value="reward.el"
              source-currency="elCurrency"
              has-higher-precision
              has-additional-selected-currency-main
              target-currency="elDisplayCurrency"
            />
          </div>
          <div>
            <h3 class="bold">
              {{ $t("dashboard.validator.blocks.cl_rewards") }}
            </h3>
            <BcFormatAmount
              :value="reward.cl"
              has-higher-precision
              has-additional-selected-currency-main
              target-currency="clDisplayCurrency"
            />
          </div>
        </div>
      </template>
    </BcTooltip>
  </div>
</template>

<style lang="scss" scoped>
.summary-reward {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--padding-small);
}
</style>
