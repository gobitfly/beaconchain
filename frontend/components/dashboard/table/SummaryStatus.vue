<script setup lang="ts">
import type { VDBSummaryStatus } from '~/types/api/validator_dashboard'
import type { SlotVizIcons } from '~/types/dashboard/slotViz'

// TODO: replace with v2.5 summary data
interface Props {
  status: VDBSummaryStatus
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const mapped = computed(() => {
  const mapCount = (count: number, key: string, icon: SlotVizIcons) => {
    return {
      icon,
      key,
      className: count > 0 ? `${key.replace('_', '-')} active` : '',
      tooltip: $t(`dashboard.validator.summary.status.${key}`, {}, count)
    }
  }

  // TODO: replace with v2.5 logic (once we got the api structs)
  const scheduledSyncCount = props.status?.next_sync_count
  const currentSyncCount = props.status?.current_sync_count
  const slashedCount = props.status?.slashed_count

  return [
    mapCount(currentSyncCount, 'current_sync', 'sync'),
    mapCount(scheduledSyncCount, 'scheduled_sync', 'sync'),
    mapCount(slashedCount, 'slashing', 'slashing')
  ]
})

</script>
<template>
  <div class="summary-status-container">
    <BcTooltip v-for="item in mapped" :key="item.key" :text="item.tooltip" :fit-content="true" class="tooltip">
      <SlotVizIcon :class="item.className" :icon="item.icon" />
    </BcTooltip>
  </div>
</template>
<style lang="scss" scoped>
@use "sass:color";
@keyframes status-rotation {
  to {
    transform: rotate(360deg);
  }
}

@mixin set-pulse-anmiation($color) {
  @keyframes pulse-animation {
  0% {
    color: color.adjust($color, $alpha: 0);
    transform: scale(0.9)
  }

  50% {
    color: color.adjust($color, $alpha: -0.4);
    transform: scale(1.1)
  }

  100% {
    color: color.adjust($color, $alpha: 0);
    transform: scale(0.9)
  }
}
}

/* unfortunatly we can't use our css variables here as the rgba conversion is done during build process and there is no css native possibility */
@include set-pulse-anmiation(#4e7451);

.dark-mode {
  @include set-pulse-anmiation(#7dc382);
}

.summary-status-container {
  background-color: var(--subcontainer-background);
  border-radius: var(--border-radius);
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  color: var(--text-color-disabled);
  height: 20px;
  padding: 0 var(--padding);
  gap: 8px;

  .tooltip {
    height: fit-content;

    svg {
      height: 12px;
      width: auto;
    }
  }

  .active {
    &.slashing {
      color: var(--negative-color);
    }

    &.current-sync {
      color: var(--positive-color);
      animation: pulse-animation 1s linear infinite;
    }

    &.scheduled-sync {
      color: var(--positive-color);
      animation: status-rotation .75s linear infinite;
    }
  }
}
</style>
