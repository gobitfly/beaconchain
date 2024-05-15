<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCheck, faClose } from '@fortawesome/pro-solid-svg-icons'

interface Props {
  available: boolean,
}
defineProps<Props>()
</script>

<template>
  <div class="check-container">
    <div v-if="available" class="checkmark">
      <FontAwesomeIcon :icon="faCheck" />
    </div>
    <div v-else class="cross">
      <FontAwesomeIcon :icon="faClose" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.check-container {
  .checkmark {
    color: var(--primary-orange);
  }

  .cross {
    color: var(--text-color-discreet);
  }
}
</style>
