<script setup lang="ts">
import {
  faPowerOff
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { ValidatorStatus } from '~/types/validator'

interface Props {
  status: ValidatorStatus,
  position?: number,
  hideLabel?: boolean
}
defineProps<Props>()

</script>
<template>
  <div class="wrapper">
    <span v-if="!hideLabel" class="status">
      {{ $t(`validator_state.${status}`) }}
      <span v-if="position"> #<BcFormatNumber :value="position" /></span>
    </span>
    <FontAwesomeIcon :icon="faPowerOff" :class="status" />
  </div>
</template>
<style lang="scss" scoped>
.wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--padding-small);
  .status{
    text-transform: capitalize;
    text-wrap: nowrap;
  }

  .online{
    color: var(--positive-color);
  }
  .deposited,
  .pending {
    color: var(--orange-color);
  }
  .exiting,
  .withdrawn,
  .offline,
  .exited,
  .slashed{
    color: var(--negative-color);
  }
}
</style>
