<script setup lang="ts">
import { type BlockStatus } from '~/types/block'
import type { TagSize, TagColor } from '~/types/tag'

interface Props {
  status?: BlockStatus,
  mobile?: boolean
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const mapped = computed(() => {
  if (!props.status) {
    return
  }
  const size: TagSize = props.mobile ? 'circle' : 'default'
  let color: TagColor
  const tStatus = $t(`block.status.${props.status}`)
  const label = props.mobile ? tStatus.substring(0, 1) : tStatus
  const tooltip = props.mobile ? tStatus : undefined
  switch (props.status) {
    case 'missed':
      color = 'failed'
      break
    case 'success':
      color = 'success'
      break
    case 'orphaned':
      color = 'orphaned'
      break
    case 'scheduled':
      color = 'dark'
      break
  }

  return {
    size,
    color,
    label,
    tooltip
  }
})

</script>
<template>
  <BcTableTag v-if="mapped" v-bind="mapped" />
</template>
