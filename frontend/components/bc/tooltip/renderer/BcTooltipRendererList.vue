<script setup lang="ts">

defineProps<{
  title?: string[],
  list?: string[][],
  note?: string[],
}>()

</script>

<template>
  <div class="renderer-list">
    <BcTooltipRendererWithBold :texts="title" />
    <ul v-if="list?.length">
      <li v-for="(item, index) in list" :key="index">
        <BcTooltipRendererWithBold :texts="item" />
      </li>
    </ul>
    <span v-if="note?.length">
      <span class="bold">{{ $t('common.note') }}: </span>
      <BcTooltipRendererWithBold :texts="note" />
    </span>
  </div>
</template>

<style scoped lang="scss">
@use "~/assets/css/fonts.scss";

.renderer-list {
  width: 220px;
  min-width: 100%;
  text-align: left;
  ul {
    padding: 0;
    margin: 0;
    padding-left: 1.4em;
  }
}
</style>
