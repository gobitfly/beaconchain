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
    <span class="positive">{{ props.success }}</span>
    <span class="slash"> / </span>
    <span :class="data.failedClass">{{ props.failed }}
      <BcFormatPercent
        class="percent"
        :base="data.sum"
        :value="props.success"
        :fixed="undefined"
        :color-break-point="80"
      />
    </span>

  </span>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.efficiency {
  @include main.truncate-text;
  display: block;

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
