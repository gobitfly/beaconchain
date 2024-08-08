<script setup lang="ts">
import type { ValidatorHistoryDuties } from '~/types/api/common'

const data: ValidatorHistoryDuties[] = [
  {
    attestation_head: { income: '123450000', status: 'success' },
    attestation_source: { income: '133450000', status: 'success' },
    attestation_target: { income: '143450000', status: 'success' },
    proposal: {
      cl_attestation_inclusion_income: '123450000',
      cl_slashing_inclusion_income: '123450000',
      cl_sync_inclusion_income: '123450000',
      el_income: '123450000',
      status: 'success',
    },
    slashing: { income: '123450000', status: 'success' },
    sync: { income: '123450000', status: 'success' },
    sync_count: 32,
  },
  {
    attestation_head: { income: '-123450000', status: 'failed' },
    attestation_source: { income: '133450000', status: 'success' },
    attestation_target: { income: '143450000', status: 'success' },
    proposal: {
      cl_attestation_inclusion_income: '123450000',
      cl_slashing_inclusion_income: '123450000',
      cl_sync_inclusion_income: '123450000',
      el_income: '123450000',
      status: 'success',
    },
    slashing: { income: '123450000', status: 'success' },
    sync: { income: '123450000', status: 'success' },
    sync_count: 32,
  },
  {
    attestation_head: { income: '123450000', status: 'success' },
    attestation_source: { income: '-133450000', status: 'failed' },
    attestation_target: { income: '143450000', status: 'success' },
    proposal: {
      cl_attestation_inclusion_income: '-133450000',
      cl_slashing_inclusion_income: '-133450000',
      cl_sync_inclusion_income: '-133450000',
      el_income: '-133450000',
      status: 'failed',
    },
    slashing: { income: '123450000', status: 'partial' },
    sync: { income: '12450000', status: 'partial' },
    sync_count: 12,
  },
  {
    attestation_head: { income: '123450000', status: 'success' },
    attestation_source: { income: '133450000', status: 'success' },
    attestation_target: { income: '-143450000', status: 'failed' },
    proposal: {
      cl_attestation_inclusion_income: '123450000',
      cl_slashing_inclusion_income: '123450000',
      cl_sync_inclusion_income: '123450000',
      el_income: '123450000',
      status: 'success',
    },
    slashing: { income: '123450000', status: 'success' },
    sync: { income: '123450000', status: 'success' },
    sync_count: 32,
  },
  {
    attestation_head: { income: '-123450000', status: 'failed' },
    attestation_source: { income: '-133450000', status: 'failed' },
    attestation_target: { income: '-143450000', status: 'failed' },
    proposal: {
      cl_attestation_inclusion_income: '-123450000',
      cl_slashing_inclusion_income: '-123450000',
      cl_sync_inclusion_income: '-123450000',
      el_income: '-123450000',
      status: 'failed',
    },
    slashing: { income: '-123450000', status: 'failed' },
    sync: { income: '-123450000', status: 'failed' },
    sync_count: 0,
  },
  {
    proposal: {
      cl_attestation_inclusion_income: '0',
      cl_slashing_inclusion_income: '0',
      cl_sync_inclusion_income: '0',
      el_income: '0',
      status: 'failed',
    },
  },
  {},
]
</script>

<template>
  <div class="container">
    <div class="item">
      <DashboardTableSummaryReward
        :reward="{ el: '213129312312', cl: '18293791237' }"
      />
    </div>
    <div class="item">
      <DashboardTableSummaryStatus
        :status="{
          next_sync_count: 123,
          current_sync_count: 0,
          slashed_count: 1,
        }"
      />
    </div>
    <div class="item">
      <DashboardTableSummaryStatus
        :status="{
          next_sync_count: 0,
          current_sync_count: 8,
          slashed_count: 0,
        }"
      />
    </div>
    <div
      v-for="(duty, index) in data"
      :key="index"
      class="item"
    >
      <ValidatorTableDutyStatus
        :data="duty"
        :compact="index === 1"
      />
      <ValidatorTableDutyRewards :data="duty" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  flex-wrap: wrap;

  .item {
    flex-grow: 0;
    display: inline-flex;
    gap: 10px;
    padding: 10px;
    margin: 10px;
    border: solid 1px beige;
    width: fit-content;
  }
}
</style>
