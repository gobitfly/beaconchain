<script setup lang="ts">
interface Props {
  alignment?: 'center' | 'default',
  hasBackdrop?: boolean,
  loading?: boolean,
  size?: 'full' | 'large' | 'medium' | 'small', // default = medium
}
defineProps<Props>()
</script>

<template>
  <div
    v-if="loading !== false"
    class="spinning-container"
    :class="{
      'center': alignment === 'center',
      'default': alignment === 'default',
      'has-backdrop': hasBackdrop,
    }"
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
  position: relative;
  z-index: 1;

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
    position: absolute;
    z-index: 2;
    width: 40px;
    height: 40px;
    vertical-align: text-bottom;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: spinner-rotation 0.75s linear infinite;

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
.has-backdrop {
    backdrop-filter: blur(2px);
  }
</style>
