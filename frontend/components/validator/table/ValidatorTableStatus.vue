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
const props = defineProps<Props>()

const iconColor = computed(() => {
  if (props.status.includes('online')) { return 'green' }
  if (props.status.includes('offline')) { return 'red' }
  return 'orange'
})

</script>
<template>
  <div class="wrapper">
    <FontAwesomeIcon :icon="faPowerOff" :class="iconColor" />
    <span v-if="!hideLabel" class="status">
      {{ $t(`validator_state.${status}`) }}
      <span v-if="position"> #<BcFormatNumber :value="position" /></span>
    </span>
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

  .green{
    color: var(--positive-color);
  }
  .red{
    color: var(--negative-color);
  }
  .orange{
    color: var(--orange-color);
  }
}
</style>
