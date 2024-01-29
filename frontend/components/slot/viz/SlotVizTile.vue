<script setup lang="ts">
import { type SlotVizSlot, type SlotVizIcons } from '~/types/dashboard/slotViz'
import { type TolltipLayout } from '~/types/layouts'
interface Props {
  data: SlotVizSlot
}
const props = defineProps<Props>()

const data = computed(() => {
  const slot = props.data
  let outer = ''
  let inner = ''
  const icons: SlotVizIcons[] = []
  switch (slot.state) {
    case 'missed':
    case 'orphaned':
      outer = 'missed'
      break
    case 'proposed':
      outer = 'proposed'
      break
  }
  const hasFaied = !!slot.duties?.find(s => s.failedCount)
  const hasSuccess = !!slot.duties?.find(s => s.successCount)
  const hasPending = !!slot.duties?.find(s => s.pendingCount)
  if (hasFaied && hasSuccess) {
    inner = 'mixed'
  } else if (hasFaied) {
    inner = 'missed'
  } else if (hasSuccess) {
    inner = 'proposed'
  } else if (hasPending) {
    inner = 'pending'
  }
  if (slot.duties?.find(s => s.type === 'propsal')) {
    icons.push('block_proposal')
  }
  if (slot.duties?.find(s => s.type === 'slashing')) {
    icons.push('slashing')
  }
  if (slot.duties?.find(s => s.type === 'sync')) {
    icons.push('sync')
  }
  if (slot.duties?.find(s => s.type === 'attestation')) {
    icons.push('attestation')
  }

  const tooltipLayout: TolltipLayout = slot.duties?.length ? 'dark' : 'default'

  return {
    id: `slot_${slot.id}`,
    tooltipLayout,
    outer,
    inner,
    icons,
    firstIconClass: `count_${icons.length}`
  }
})

</script>
<template>
  <BcTooltip :target="data.id" :layout="data.tooltipLayout">
    <div :id="data.id" class="tile" :class="data.outer">
      <div class="inner" :class="data.inner">
        <IconPlus v-show="data.icons?.length > 2" class="plus" />
        <SlotVizIcon v-if="data.icons?.length" :icon="data.icons[0]" class="first_icon" :class="data.firstIconClass" />
        <SlotVizIcon v-if="data.icons?.length === 2" :icon="data.icons[1]" class="second_icon" />
      </div>
    </div>
    <template #tooltip>
      <div class="tooltip-content">
        <h3>My Tooltip</h3>
        <p>This is a tooltip with complex content.</p>
        <Button label="Click Me" />
      </div>
    </template>
  </BcTooltip>
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
    left: 3px;
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
