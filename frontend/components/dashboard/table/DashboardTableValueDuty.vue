<script setup lang="ts">
import type { VDBRewardsTableDuty } from '~/types/api/validator_dashboard'

interface Props {
  duty: VDBRewardsTableDuty,
  isNumberVisible?: boolean,
}
withDefaults(
  defineProps<Props>(),
  {
    isNumberVisible: true,
  },
)
</script>

<template>
  <div
    v-if="duty"
    class="duty"
  >
    <div
      v-if="duty.attestation !== undefined"
      class="group"
    >
      {{ $t("dashboard.validator.rewards.attestation") }}
      <BcFormatPercent
        v-if="isNumberVisible"
        class="round-brackets"
        :percent="duty.attestation"
        :maximum-fraction-digits="0"
      />
    </div>
    <div
      v-if="duty.proposal !== undefined"
      class="group"
    >
      {{ $t("dashboard.validator.rewards.proposal") }}
      <BcFormatPercent
        v-if="isNumberVisible"
        class="round-brackets"
        :percent="duty.proposal"
        :maximum-fraction-digits="0"
      />
    </div>
    <div
      v-if="duty.sync !== undefined"
      class="group"
    >
      {{ $t("dashboard.validator.rewards.sync_committee") }}
      <BcFormatPercent
        v-if="isNumberVisible"
        class="round-brackets"
        :percent="duty.sync"
        :maximum-fraction-digits="0"
      />
    </div>
    <div
      v-if="duty.slashing !== undefined"
      class="group"
    >
      {{ $t("dashboard.validator.rewards.slashing") }}
      <BcFormatNumber
        v-if="isNumberVisible"
        class="round-brackets"
        :value="duty.slashing"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.duty {
  display: inline-flex;
  flex-wrap: wrap;

  .group {
    white-space: nowrap;
    text-wrap: nowrap;

    &:not(:last-child) {
      &:after {
        content: ",\00a0";
      }
    }
  }
}
</style>
