<script setup lang="ts">
import type { ClElValue } from '~/types/api/common'

interface Props {
  reward?: ClElValue<string>,
  status?: 'success' | 'missed' | 'orphaned' | 'scheduled'
}
defineProps<Props>()

</script>
<template>
  <BcTooltip v-if="status === 'success' && reward" class="combine-rewards">
    <BcFormatValue :value="reward?.el" :no-tooltip="true" />
    <BcFormatValue v-if="reward?.cl" :value="reward?.cl" :no-tooltip="true" />
    <span v-else>{{ $t('dashboard.validator.blocks.cl_pending') }}</span>
    <template #tooltip>
      <div>
        <div class="tt-row">
          <span>{{ $t('dashboard.validator.blocks.el_rewards') }}: </span>
          <BcFormatValue :value="reward?.el" :no-tooltip="true" :full-value="true" />
        </div>
        <div class="tt-row">
          <span>{{ $t('dashboard.validator.blocks.cl_rewards') }}: </span>
          <BcFormatValue v-if="reward?.cl" :value="reward?.cl" :no-tooltip="true" :full-value="true" />
          <span v-else>{{ $t('dashboard.validator.blocks.pending') }}</span>
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

  >div:last-child {
    font-size: var(--small_text_font_size);
    color: var(--text-color-discreet);
  }
}
</style>
