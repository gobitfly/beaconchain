<script setup lang="ts">
import {
  faPowerOff
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { ValidatorStatus } from '~/types/validator'

interface Props {
  status: ValidatorStatus,
  position?: number
}
const props = defineProps<Props>()

</script>
<template>
  <div class="wrapper">
    <span class="status">
      {{ $t(`validator_state.${props.status}`) }}
      <span v-if="position"> #{{ position }}</span>
    </span>
    <FontAwesomeIcon :icon="faPowerOff" :class="status" />
  </div>
</template>
<style lang="scss" scoped>
.wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .status{
    text-transform: capitalize;
  }

  .active{
    color: var(--positive-color);
  }
  .deposited,
  .pending {
    color: var(--orange-color);
  }
  .inactive,
  .exiting,
  .exited,
  .slashed{
    color: var(--negative-color);
  }
}
</style>
