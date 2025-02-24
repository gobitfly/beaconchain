<script setup lang="ts">
import type { ClElValue } from '~/types/api/common'

defineProps<{
  reward?: ClElValue<string>,
  status?: 'missed' | 'orphaned' | 'scheduled' | 'success',
}>()
</script>

<template>
  <BcTooltip
    v-if="status === 'success' && reward"
    class="combine-rewards"
    fit-content
  >
    <BcFormatAmount
      :value="reward?.el"
      source-currency="elCurrency"
      target-currency="elDisplayCurrency"
    />
    <BcFormatAmount
      v-if="reward?.cl && reward.cl != '0'"
      :value="reward?.cl"
      source-currency="clCurrency"
      target-currency="clDisplayCurrency"
    />
    <span v-else>{{ $t("dashboard.validator.blocks.cl_pending") }}</span>
    <template #tooltip>
      <div>
        <div class="tt-row">
          <span>{{ $t("dashboard.validator.blocks.el_rewards") }}: </span>
          <BcFormatAmount
            :value="reward?.el"
            source-currency="elCurrency"
            target-currency="elDisplayCurrency"
            has-additional-selected-currency-main
            :maximum-fraction-digits="unitFactorCrypto.base"
          />
        </div>
        <div class="tt-row">
          <span>{{ $t("dashboard.validator.blocks.cl_rewards") }}: </span>
          <template
            v-if="reward?.cl && reward.cl != '0'"
          >
            <BcFormatAmount
              :value="reward?.cl"
              target-currency="clDisplayCurrency"
              has-additional-selected-currency-main
              :maximum-fraction-digits="unitFactorCrypto.base"
            />
          </template>
          <span v-else>{{ $t("dashboard.validator.blocks.pending") }}</span>
        </div>
      </div>
    </template>
  </BcTooltip>
  <span v-else>-</span>
</template>

<style lang="scss" scoped>
.tt-row {
  display: flex;
  flex-wrap: nowrap;
  white-space: nowrap;
  gap: 3px;
}

.combine-rewards {
  display: inline-flex;
  flex-direction: column;

  > div:last-child,
  > span:last-child {
    font-size: var(--small_text_font_size);
    color: var(--text-color-discreet);
  }
}
</style>
