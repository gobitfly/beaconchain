<script setup lang="ts">

interface Props {
  success: number,
  failed: number,
}
const props = defineProps<Props>()

const data = computed(() => {
  const failedClass = props.failed ? 'negative' : 'positive'
  const sum = props.failed + props.success

  return { failedClass, sum }
})

</script>
<template>
  <span class="efficiency">
    <BcFormatNumber class="positive" :value="props.success " />
    <span class="slash"> / </span>
    <BcFormatNumber :class="data.failedClass" :value="props.failed " /> <BcFormatPercent
      class="percent"
      :base="data.sum"
      :value="props.success"
      :fixed="undefined"
      :color-break-point="80"
      :full-on-empty-base="true"
    />
  </span>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.efficiency {

  .positive {
    color: var(--positive-color);
  }

  .negative {
    color: var(--negative-color);
  }

  .percent {
    &::before {
      content: "(";
    }

    &::after {
      content: ")";
    }
  }
}
</style>
