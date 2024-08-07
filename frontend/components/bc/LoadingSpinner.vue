<script setup lang="ts">
interface Props {
  loading?: boolean
  size?: 'small' | 'medium' | 'large' | 'full' // default = medium
  alignment?: 'center' | 'default'
}
defineProps<Props>()
</script>

<template>
  <div
    v-if="loading !== false"
    class="spinning-container"
    :class="[alignment]"
  >
    <div
      class="spinner"
      :class="[size]"
    >
      <span />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@keyframes spinner-rotation {
  to {
    transform: rotate(360deg);
  }
}

.spinning-container {
  color: var(--primary-color);

  &.center {
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    flex-grow: 1;
  }

  &:has(.full) {
    width: 100%;
    height: 100%;
  }

  .spinner {
    display: inline-block;
    width: 40px;
    height: 40px;
    vertical-align: text-bottom;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: spinner-rotation .75s linear infinite;

    &.small {
      width: 20px;
      height: 20px;
      border-width: 1px;
    }

    &.large {
      width: 80px;
      height: 80px;
      border-width: 4px;
    }

    &.full {
      width: 100%;
      height: 100%;
    }

  }

}
</style>
