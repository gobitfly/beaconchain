<script setup lang="ts">
const HIGLIGHT_KEY = ['highlight', 'highlight_1', 'highlight_2', 'highlight_3'] // possible highlight key's in the translation
defineProps<{
  tPath: string, // Path to the translation
  tHighlightPath?: string, // Path to the translation highlight definitions
  tOptions?: any // translation options
  tag?: string // html tag to wrap the translations
  url?: string // url for a url placeholder
  target?: string // url target - defaults to _blank
}>()

</script>

<template>
  <i18n-t :keypath="tPath" :tag="tag || 'span'" :plural="tOptions?.plural">
    <!-- render highlight placeholders -->
    <template v-for="(hKey, hIndex) in HIGLIGHT_KEY" :key="hKey" #[hKey]>
      <BcTooltipRendererText
        v-if="tHighlightPath"
        class="bold"
        :t-path="`${tHighlightPath}${hIndex ? `[${hIndex-1}]`:''}`"
        :t-options="tOptions"
        :url="url"
        :target="target"
      />
    </template>
    <!-- render url -->
    <template v-if="url" #url>
      <BcLink :to="url" :target="target ? target : '_blank'" class="link" />
    </template>
    <!-- values from the options -->
    <template v-for="(value, oKey) in tOptions" :key="oKey" #[oKey]>
      <span v-if="value && `${oKey}` !== 'plural'">{{ value }}</span>
    </template>
  </i18n-t>
</template>
