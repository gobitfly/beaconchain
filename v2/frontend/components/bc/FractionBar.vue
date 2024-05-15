<script lang="ts" setup>
interface Props {
  barFillPercentage?: number,
}
const props = defineProps<Props>()

const fillPercentage = computed(() => {
  return props.barFillPercentage ? `${props.barFillPercentage}%` : '0'
})
</script>

<template>
  <div v-if="barFillPercentage" class="fraction-bar">
    <div class="fill" :style="{ width: fillPercentage }" />
  </div>
</template>

<style lang="scss" scoped>
.fraction-bar {
  background-color: var(--dark-grey);
  border-radius: 5px;
  position: relative;

  .fill {
    height: 100%;
    background-color: var(--primary-orange);
    border-radius: 5px;
    position: absolute;
  }
}
</style>
