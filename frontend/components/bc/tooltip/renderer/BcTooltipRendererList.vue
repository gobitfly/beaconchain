<script setup lang="ts">

defineProps<{
  tKeys?: string[],
  tHighlightPath?: string,
  tOptions?: any
  url?: string
  target?: string
}>()

</script>

<template>
  <ul v-if="tKeys?.length" class="renderer-list">
    <BcTooltipRendererText
      v-for="(path, index) in tKeys"
      :key="index"
      :t-path="path"
      :t-highlight-path="tHighlightPath"
      :t-options="tOptions"
      :url="url"
      :target="target"
      tag="li"
    />
  </ul>
</template>

<style scoped lang="scss">
.renderer-list {
  text-align: left;
  padding: 0;
  margin: 0;
  padding-left: 20px;
}
</style>
