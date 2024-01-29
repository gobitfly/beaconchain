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
          {{ row.state === 'head' ? $t('slotViz.head') : row.id }}
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
.content {
  display: flex;
  gap: var(--padding);
  overflow-x: auto;
  width: 100%;
  height: 220px;

  .rows {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--padding-large);
    height: 30px;

    .row {
      display: flex;
      align-items: center;
      gap: var(--padding);
    }
  }
}
</style>
