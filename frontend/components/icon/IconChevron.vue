<script setup lang="ts">
interface Props {
  direction?: 'left'| 'right'| 'top' | 'bottom'
}
const props = defineProps<Props>()
</script>
<template>
  <svg
    class="chevron"
    :class="props.direction"
    xmlns="http://www.w3.org/2000/svg"
    width="14"
    height="10"
    viewBox="0 0 14 10"
    fill="none"
  >
    <path
      d="M7.81923 8.82967C7.42113 9.39838 6.57887 9.39838 6.18077 8.82967L1.10143 1.57346C0.637482 0.910686 1.11163 -2.85415e-07 1.92066 -2.07635e-07L12.0793 7.69022e-07C12.8884 8.46802e-07 13.3625 0.910686 12.8986 1.57346L7.81923 8.82967Z"
      fill="currentColor"
    />
  </svg>
</template>
<style lang="scss" scoped>
.chevron {
  transform: rotate(-90deg);
  transition: transform 0.2s;

  &.bottom {
    transform: rotate(0);
  }
  &.left {
    transform: rotate(90deg);
  }

  &.top {
    transform: rotate(180deg);
  }
}
</style>
