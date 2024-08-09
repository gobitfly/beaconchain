<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faCheck, faClose,
} from '@fortawesome/pro-solid-svg-icons'

interface Props {
  available?: boolean
}
defineProps<Props>()
</script>

<template>
  <FontAwesomeIcon
    v-if="available"
    :icon="faCheck"
    class="checkmark"
  />
  <FontAwesomeIcon
    v-else
    :icon="faClose"
    class="cross"
  />
</template>

<style lang="scss" scoped>
.checkmark {
  color: var(--primary-orange);
  width: 18px;
  height: 20px;
}

.cross {
  color: var(--text-color-discreet);
  width: 14px;
  height: 20px;
}
</style>
