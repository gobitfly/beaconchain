<script setup lang="ts">
import type { SummaryValidatorsIconRowInfo } from '~/types/validator'

interface Props {
  absolute: boolean,
  icons: SummaryValidatorsIconRowInfo[],
  total?: number,
}
const props = defineProps<Props>()

const combinedTotal = computed<number>(
  () =>
    props.total ?? props.icons?.reduce((sum, icon) => sum + icon.count, 0) ?? 0,
)
</script>

<template>
  <div
    v-for="status in icons"
    :key="status.key"
    class="status"
    :class="status.key"
  >
    <div class="icon">
      <BcIcon name="power-off" />
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
@use "~/assets/css/utils.scss";

.status {
  display: flex;
  align-items: center;
  gap: 3px;

  .icon {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-color-disabled);

  }

  &.online {
    .icon {
      color: var(--positive-color);
    }

    span {
      color: var(--positive-color);
    }
  }

  &.offline {
    .icon {
      color: var(--negative-color);
    }

    span {
      color: var(--negative-color);
    }
  }
}
</style>
