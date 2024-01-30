<script setup lang="ts">
import { type SlotVizData } from '~/types/dashboard/slotViz'
interface Props {
  data: SlotVizData
}
const props = defineProps<Props>()

const rows = computed(() => {
  const data = props.data
  // Todo: add empty placeholder data while loading
  return data.epochs
})

</script>
<template>
  <div class="content">
    <div class="rows">
      <div v-for="row in rows" :key="row.id" class="row">
        <div class="epoch">
          {{ row.state === 'head' ? $t('slotViz.head') : row.id.toLocaleString('en-US') }}
        </div>
      </div>
    </div>
    <div class="rows">
      <div v-for="row in rows" :key="row.id" class="row">
        <SlotVizTile v-for="slot in row.slots" :key="slot.id" :data="slot" />
      </div>
    </div>
  </div>
</template>
<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.content {
  @include main.container;
  display: flex;
  gap: var(--padding);
  overflow-x: auto;
  overflow-y: hidden;
  min-height: 180px;
  min-height: 180px;
  padding: var(--padding-large) var(--padding-large) var(--padding-large) 9px;

  .epoch {
    @include fonts.small_text;
  }

  .rows {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--padding-large);

    .row {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      height: 30px;
      gap: var(--padding);
    }
  }
}
</style>
