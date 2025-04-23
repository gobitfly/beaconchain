<script setup lang="ts">
import BcIcon from '~/components/bc/icon/BcIcon.vue'
import type { ValidatorStatus } from '~/types/validator'

interface Props {
  hideLabel?: boolean,
  status: ValidatorStatus,
}
const props = defineProps<Props>()

const iconColor = computed(() => {
  if (props.status.includes('online')) return 'green'
  if (props.status.includes('offline')) return 'red'
  return 'orange'
})
</script>

<template>
  <div class="wrapper">
    <BcIcon
      name="power-off"
      :class="iconColor"
    />
    <span
      v-if="!hideLabel"
      class="status"
    >
      {{ $t(`validator_state.${status}`) }}
    </span>
  </div>
</template>

<style lang="scss" scoped>
.wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: var(--padding-small);
  .status {
    text-transform: capitalize;
    text-wrap: nowrap;
  }

  .green {
    color: var(--positive-color);
  }
  .red {
    color: var(--negative-color);
  }
  .orange {
    color: var(--orange-color);
  }
}
</style>
