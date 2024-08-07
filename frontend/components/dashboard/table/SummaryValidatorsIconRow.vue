<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faPowerOff,
} from '@fortawesome/pro-solid-svg-icons'
import type { SummaryValidatorsIconRowInfo } from '~/types/validator'

interface Props {
  icons: SummaryValidatorsIconRowInfo[]
  total?: number
  absolute: boolean
}
const props = defineProps<Props>()

const combinedTotal = computed<number>(() => props.total ?? props.icons?.reduce((sum, icon) => sum + icon.count, 0) ?? 0)
</script>

<template>
  <div
    v-for="status in icons"
    :key="status.key"
    class="status"
    :class="status.key"
  >
    <div class="icon">
      <FontAwesomeIcon :icon="faPowerOff" />
    </div>
    <BcFormatNumber
      v-if="absolute"
      :value="status.count"
    />
    <BcFormatPercent
      v-else
      :value="status.count"
      :base="combinedTotal"
    />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.status {
  display: flex;
  align-items: center;
  gap: 3px;

  .icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background-color: var(--text-color-disabled);

    svg {
      height: 8px;
      width: 8px;
    }
  }

  &.online {
    .icon {
      background-color: var(--positive-color);
      color: var(--positive-contrast-color);
    }

    span {
      color: var(--positive-color);
    }
  }

  &.offline {
    .icon {
      background-color: var(--negative-color);
      color: var(--negative-contrast-color);
    }

    span {

      color: var(--negative-color);
    }
  }
}
</style>
