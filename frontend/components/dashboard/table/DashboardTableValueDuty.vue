<script setup lang="ts">
import type { VDBRewardesTableDuty } from '~/types/api/validator_dashboard'

interface Props {
  duty: VDBRewardesTableDuty;
}
defineProps<Props>()

</script>
<template>
  <div v-if="duty" class="duty">
    <div class="group">
      {{ $t('dashboard.validator.rewards.attestation') }}
      <BcFormatPercent class="round-brackets" :percent="duty.attestation" :fixed="0" :precision="0" />,
    </div>
    <div class="group">
      {{ $t('dashboard.validator.rewards.proposal') }}
      <BcFormatPercent class="round-brackets" :percent="duty.proposal" :fixed="0" :precision="0" />,
    </div>
    <div class="group">
      {{ $t('dashboard.validator.rewards.sync_committee') }}
      <BcFormatPercent class="round-brackets" :percent="duty.sync" :fixed="0" :precision="0" />,
    </div>
    <div class="group">
      {{ $t('dashboard.validator.rewards.slashing') }}
      <BcFormatNumber class="round-brackets" :value="duty.slashing" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.duty {
  display: inline-flex;
  flex-wrap: wrap;

  .group {
    white-space: nowrap;
    text-wrap: nowrap;

    &:not(:last-child) {
      &:after {
        content: "\00a0";
      }
    }
  }
}
</style>
