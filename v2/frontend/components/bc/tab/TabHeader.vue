<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

interface Props {
  header?: string,
  icon?: IconDefinition
}

const props = defineProps<Props>()

</script>

<template>
  <div class="tab-header">
    <slot name="icon">
      <FontAwesomeIcon v-if="props.icon" :icon="props.icon" />
    </slot>
    <span v-if="props.header">{{ props.header }}</span>
  </div>
</template>

<style lang="scss" scoped>
.tab-header {
  display: inline-flex;
  align-items: center;
  gap: var(--padding-small);
}
</style>
