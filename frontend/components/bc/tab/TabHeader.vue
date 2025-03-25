<script setup lang="ts">
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

interface Props {
  header?: string,
  icon?: Icon,
}

const props = defineProps<Props>()
</script>

<template>
  <div class="tab-header">
    <slot name="icon">
      <BcIcon
        v-if="props.icon"
        :name="props.icon"
      />
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
