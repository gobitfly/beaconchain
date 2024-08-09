<script setup lang="ts">
import { type BlockStatus } from '~/types/block'
import type {
  TagColor, TagSize,
} from '~/types/tag'

interface Props {
  blockSlot?: number
  mobile?: boolean
  status?: BlockStatus
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()

const { latestState } = useLatestStateStore()

// we don't want to be reactive to the current_slot
const currentSlot = latestState.value?.current_slot || 0

const mapped = computed(() => {
  if (!props.status) {
    return
  }
  const size: TagSize = props.mobile ? 'circle' : 'default'
  let color: TagColor
  const status
    = props.status === 'scheduled'
    && props.blockSlot
    && props.blockSlot < currentSlot
      ? 'probably_missed'
      : props.status
  const tStatus = $t(`block.status.${status}`)
  const label = props.mobile ? tStatus.substring(0, 1) : tStatus
  const tooltip
    = status === 'probably_missed'
      ? $t('block.status_might_change_on_reorg')
      : props.mobile
        ? tStatus
        : undefined
  switch (status) {
    case 'missed':
      color = 'failed'
      break
    case 'orphaned':
      color = 'orphaned'
      break
    case 'probably_missed':
      color = 'partial'
      break
    case 'scheduled':
      color = 'dark'
      break
    case 'success':
      color = 'success'
      break
  }

  return {
    color,
    label,
    size,
    status,
    tooltip,
  }
})
</script>

<template>
  <BcTableTag
    v-if="mapped"
    v-bind="mapped"
  />
</template>
