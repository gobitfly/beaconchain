<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ClElValue } from '~/types/api/common'

interface Props {
  reward?: ClElValue<string>
}
const props = defineProps<Props>()

const total = computed(() =>
  props.reward ? totalElCl(props.reward) : undefined,
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
      :use-colors="true"
    />
    <BcTooltip :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.el_rewards") }}:
            </span>
            <BcFormatValue
              :value="reward?.el"
              :no-tooltip="true"
              :full-value="true"
            />
          </div>
          <div class="tt-row">
            <span class="bold">{{ $t("dashboard.validator.blocks.cl_rewards") }}:
            </span>
            <BcFormatValue
              :value="reward?.cl"
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
}

.summary-reward {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--padding-small);
}
</style>
