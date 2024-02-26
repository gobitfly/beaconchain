<script setup lang="ts">
import type { VDBSlotVizSlot } from '~/types/api/slot_viz'
import { type SlotVizIcons } from '~/types/dashboard/slotViz'
interface Props {
  data: VDBSlotVizSlot
}
const props = defineProps<Props>()

const data = computed(() => {
  const slot = props.data
  let outer = ''
  const icons: SlotVizIcons[] = []
  switch (slot.status) {
    case 'missed':
    case 'orphaned':
      outer = 'missed'
      break
    case 'proposed':
      outer = 'proposed'
      break
  }

  let inner = ''
  if (slot.status === 'scheduled') {
    inner = 'pending'
  } else {
    const hasFailed = !!slot.attestations?.failed_count || !!slot.sync?.failed_count || [...slot.proposals ?? [], ...slot.slashing ?? []].find(duty => duty.status === 'failed')
    const hasSuccess = !!slot.attestations?.success_count || !!slot.sync?.success_count || [...slot.proposals ?? [], ...slot.slashing ?? []].find(duty => duty.status === 'success')
    const hasPending = !!slot.attestations?.pending_count || !!slot.sync?.pending_count || [...slot.proposals ?? [], ...slot.slashing ?? []].find(duty => duty.status === 'scheduled')
    if (!hasFailed && !hasSuccess && !hasPending) {
      inner = 'proposed'
    } else if (hasFailed && hasSuccess) {
      inner = 'mixed'
    } else if (hasFailed) {
      inner = 'missed'
    } else if (hasSuccess) {
      inner = 'proposed'
    } else if (hasPending) {
      inner = 'pending'
    }
  }

  if (slot.proposals?.length) {
    icons.push('proposal')
  }
  if (slot.slashing?.length) {
    icons.push('slashing')
  }
  if (slot.sync) {
    icons.push('sync')
  }
  if (slot.attestations) {
    icons.push('attestation')
  }

  return {
    id: `slot_${slot.slot}`,
    outer,
    inner,
    icons,
    firstIconClass: `count_${icons.length}`
  }
})
</script>
<template>
  <SlotVizTooltip :id="data.id" :data="props.data">
    <div :id="data.id" class="tile" :class="data.outer">
      <div class="inner" :class="data.inner">
        <IconPlus v-show="data.icons?.length > 2" class="plus" />
        <SlotVizIcon v-if="data.icons?.length" :icon="data.icons[0]" class="first_icon" :class="data.firstIconClass" />
        <SlotVizIcon v-if="data.icons?.length === 2" :icon="data.icons[1]" class="second_icon" />
      </div>
    </div>
  </SlotVizTooltip>
</template>
<style lang="scss" scoped>
.tile {
  display: flex;
  width: 30px;
  height: 30px;
  background-color: var(--asphalt);
}

.inner {
  position: relative;
  width: 24px;
  height: 24px;
  margin: 3px;
  background-color: inherit;
  color: var(--light-grey);

  .plus {
    position: absolute;
    top: 2px;
    right: 2px;
    width: 11px;
    height: 11px;
  }

  .first_icon {
    position: absolute;
    top: 10px;
    left: 1px;
    width: 16px;
    height: 14px;

    &.count_1 {
      top: 3px;
      left: 3px;
      width: 19px;
      height: 18px;
    }

    &.count_2 {
      top: 10px;
      left: 10px;
      width: 16px;
      height: 16px;
    }
  }

  .second_icon {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 11px;
    height: 10px;
  }
}

.pending {
  background-color: var(--asphalt);

  &.inner {
    color: var(--light-grey);
  }
}

.proposed {
  background-color: var(--flashy-green);

  &.inner {
    color: var(--graphite);
  }
}

.missed {
  background-color: var(--flashy-red);

  &.inner {
    color: var(--graphite);
  }
}

.mixed {
  background-color: var(--yellow);

  &.inner {
    color: var(--graphite);
  }
}
</style>
