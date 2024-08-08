<script setup lang="ts">
import type { TagColor } from '~/types/tag'

interface Props {
  valid?: boolean
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()

const data = computed<{ color: TagColor, label: string }>(() => {
  return {
    color: props.valid ? 'success' : 'failed',
    label: props.valid ? $t('common.true') : $t('common.false'),
  }
})
</script>

<template>
  <BcTableTag
    :color="data.color"
    :label="data.label"
    size="compact"
  />
</template>
