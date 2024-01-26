<script setup lang="ts">
import { type SlotVizSlot } from '~/types/dashboard/slotViz'
interface Props {
  data: SlotVizSlot
}
const props = defineProps<Props>()

const data = computed(() => {
  console.log(props)
  return {
    outer: '',
    inner: '',
    icons: []
  }
})

</script>
<template>
  <div class="tile" :class="data.outer">
    <div class="inner" :class="data.outer">
      <SlotVizIcon v-for="icon in data.icons" :key="icon" :icon="icon" />
    </div>
  </div>
</template>
<style lang="scss" scoped>
.tile {
  with: 30px;
  height: 30px;

  .inner {
    with: 26px;
    height: 26px;
    margin: 3px;
  }
}
</style>
