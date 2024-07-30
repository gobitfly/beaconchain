<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'

const props = defineProps<{
  tPath: string,
  tOptions?: any,
}>()

const { t: $t } = useI18n()

const keys = computed(() => {
  const title = hasTranslation($t, `${props.tPath}.title`) ? `${props.tPath}.title` : undefined

  const list: string[] = []
  let notFound = false
  while (!notFound) {
    const path: string = `${props.tPath}.list.${(list.length)}`
    if (hasTranslation($t, path)) {
      list.push(path)
    } else {
      notFound = true
    }
  }

  const note = hasTranslation($t, `${props.tPath}.note`) ? `${props.tPath}.note` : undefined
  return {
    title,
    titleHighlight: `${props.tPath}.highlight.title`,
    list,
    listHighlight: `${props.tPath}.highlight.list`,
    note,
    noteHighlight: `${props.tPath}.highlight.note`,
    hasItems: title || list.length || note
  }
})

</script>

<template>
  <BcTooltip v-if="keys.hasItems" :fit-content="true">
    <FontAwesomeIcon :icon="faInfoCircle" class="info" />
    <template #tooltip>
      <div class="list-tooltip">
        <BcTooltipRendererText
          v-if="keys.title"
          :t-path="keys.title"
          :t-highlight-path="keys.titleHighlight"
          :t-options="tOptions"
        />
        <BcTooltipRendererList v-if="keys.list.length" :t-keys="keys.list" :t-options="tOptions" />
        <BcTooltipRendererText
          v-if="keys.note"
          :t-path="keys.note"
          :t-highlight-path="keys.noteHighlight"
          :t-options="tOptions"
        />
      </div>
    </template>
  </BcTooltip>
</template>

<style scoped lang="scss">
.list-tooltip {
  text-align: left;
  width: 220px;
  min-width: 100%;
}
</style>
