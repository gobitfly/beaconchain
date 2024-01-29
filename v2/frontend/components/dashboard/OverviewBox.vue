<script setup lang="ts">
import { type OverviewTableData } from '~/types/dashboard/overview'
interface Props {
  data: OverviewTableData
}
const props = defineProps<Props>()

const value = computed(() => props.data.value ?? '-')

</script>
<template>
  <div class="box">
    <div class="main">
      <div class="big_text_label">
        {{ props.data.label }}
      </div>
      <div class="big_text">
        {{ value }}
      </div>
    </div>
    <div v-for="(infos, index) in props.data.additonalValues" :key="index" class="additional">
      <div v-for="addValue in infos" :key="addValue" class="small_text">
        {{ addValue }}
      </div>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.box {
  display: flex;
  align-items: center;

  .main,
  .additional {
    display: flex;
    flex-direction: column;

    div {
      display: inline-block;
      white-space: nowrap;
      text-wrap: nowrap;
    }
  }

  .additional {
    margin-left: 8px;

    &:nth-child(2) {
      margin-left: var(--padding);
    }
  }

}
</style>
