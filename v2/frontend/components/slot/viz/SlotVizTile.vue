<script setup lang="ts">
import type { VDBSlotVizSlot } from '~/types/api/slot_viz'
import type {
  SlotVizCategories,
  SlotVizIcons,
} from '~/types/dashboard/slotViz'

interface Props {
  currentSlotId?: number,
  data: VDBSlotVizSlot,
  selectedCategories: SlotVizCategories[],
}
const props = defineProps<Props>()

const data = computed(() => {
  const slot: VDBSlotVizSlot = {
    ...props.data,
    attestations: props.selectedCategories.includes('attestation')
      ? props.data.attestations
      : undefined,
    proposal: props.selectedCategories.includes('proposal')
      ? props.data.proposal
      : undefined,
    slashing: props.selectedCategories.includes('slashing')
      ? props.data.slashing
      : undefined,
    sync: props.selectedCategories.includes('sync')
      ? props.data.sync
      : undefined,
  }
  const showIcons = props.selectedCategories.includes('visible')
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
    case 'scheduled':
      if (props.currentSlotId && props.currentSlotId > slot.slot) {
        outer = 'scheduled-past'
      }
      else if (props.currentSlotId === slot.slot) {
        outer = 'scheduled-current'
      }
      break
  }
  if (slot.slot === props.currentSlotId) {
    outer += ' blinking-animation'
  }

  let inner = ''
  if (slot.status === 'scheduled') {
    inner = 'pending'
  }
  else {
    const hasFailed
      = !!slot.attestations?.failed
      || !!slot.sync?.failed
      || !!slot.slashing?.failed
      || (!!slot.proposal && slot.status === 'missed')
    const hasSuccess
      = !!slot.attestations?.success
      || !!slot.sync?.success
      || !!slot.slashing?.success
      || (!!slot.proposal && slot.status === 'proposed')
    const hasPending = !!slot.attestations?.scheduled || !!slot.sync?.scheduled
    if (!hasFailed && !hasSuccess && !hasPending) {
      inner = 'proposed'
    }
    else if (hasFailed && hasSuccess) {
      inner = 'mixed'
    }
    else if (hasFailed) {
      inner = 'missed'
    }
    else if (hasSuccess) {
      inner = 'proposed'
    }
    else if (hasPending) {
      inner = 'pending'
    }
  }

  if (showIcons) {
    if (slot.proposal) {
      icons.push('proposal')
    }
    if (slot.slashing) {
      icons.push('slashing')
    }
    if (slot.sync) {
      icons.push('sync')
    }
    if (slot.attestations) {
      icons.push('attestation')
    }
  }

  return {
    firstIconClass: `count_${icons.length}`,
    icons,
    id: `slot_${slot.slot}`,
    inner,
    outer,
  }
})
</script>

<template>
  <SlotVizTooltip
    :id="data.id"
    :data="props.data"
    :current-slot-id
  >
    <div
      :id="data.id"
      class="tile"
      :class="data.outer"
    >
      <div
        class="inner"
        :class="data.inner"
      >
        <IconPlus
          v-show="data.icons?.length > 2"
          class="plus"
        />
        <SlotVizIcon
          v-if="data.icons?.length"
          :icon="data.icons[0]"
          class="first_icon"
          :class="data.firstIconClass"
        />
        <SlotVizIcon
          v-if="data.icons?.length === 2"
          :icon="data.icons[1]"
          class="second_icon"
        />
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
.scheduled-current {
  box-shadow: 0px 0px 2px var(--text-color);
}
.scheduled-past {
  opacity: 0.5;
  background-color: rgb(100, 100, 100);
  .inner {
    background-color: rgb(100, 100, 100);
  }
}

.pending {
  background-color: var(--asphalt);

  &.inner {
    color: var(--light-grey);
  }
}

.proposed {
  background-color: var(--green);

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
