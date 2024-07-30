<script setup lang="ts">
const HIGLIGHT_KEY = ['highlight', 'highlight_1', 'highlight_2', 'highlight_3'] // possible highlight key's in the translation
defineProps<{
  tPath: string,
  tHighlightPath?: string,
  tOptions?: any
  tag?: string
  url?: string
  target?: string
}>()

</script>

<template>
  <i18n-t :keypath="tPath" :tag="tag || 'span'" :plural="tOptions?.plural">
    <!-- render highlight placeholders -->
    <template v-for="(key, index) in HIGLIGHT_KEY" :key="key" #[key]>
      <BcTooltipRendererText
        v-if="tHighlightPath"
        class="bold"
        :t-path="`${tHighlightPath}${index ? `[${index-1}]`:''}`"
        :t-options="tOptions"
        :url="url"
        :target="target"
      />
    </template>
    <!-- render url -->
    <template v-if="url" #url>
      <BcLink :to="url" :target="target ?? '_blank'" class="link" />
    </template>
    <!-- values from the options -->
    <template v-for="(value, key) in tOptions" :key="key" #[key]>
      <span v-if="value && `${key}` !== 'plural'">{{ value }}</span>
    </template>
  </i18n-t>
</template>
