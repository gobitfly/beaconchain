<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'

const props = defineProps<{
  tPath: string,
  tOptions?: any,
}>()

const { t: $t } = useI18n()

const texts = computed(() => {
  const title = tAll($t, props.tPath + '.title', props.tOptions)

  const list = []
  let notFound = true
  while (notFound) {
    const path: string = `${props.tPath}.list.${(list.length)}`
    const items = tAll($t, `${path}`, props.tOptions)
    if (!items.length) {
      notFound = false
    } else {
      list.push(items)
    }
  }

  const note = tAll($t, props.tPath + '.note', props.tOptions)
  return {
    title,
    list,
    note,
    hasItems: title.length || list.length || note.length
  }
})

</script>

<template>
  <BcTooltip v-if="texts.hasItems" :fit-content="true">
    <FontAwesomeIcon :icon="faInfoCircle" class="info" />
    <template #tooltip>
      <BcTooltipRendererList :title="texts.title" :list="texts.list" :note="texts.note" />
    </template>
  </BcTooltip>
</template>
