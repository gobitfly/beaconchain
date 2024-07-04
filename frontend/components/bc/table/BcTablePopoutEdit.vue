<script setup lang="ts">
import {
  faEdit
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const emit = defineEmits<{(e: 'onEdit'): void }>()

interface Props {
  label?: string,
  noIcon?: boolean
}
defineProps<Props>()

</script>
<template>
  <div class="bc-poput-edit">
    <slot name="content">
      <span v-if="label" class="content">
        {{ label }}
      </span>
    </slot>
    <FontAwesomeIcon v-if="!noIcon" class="link" :icon="faEdit" @click="() => emit('onEdit')" />
  </div>
</template>
<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.bc-poput-edit {
  display: flex;
  align-items: center;

  .content {
    @include utils.truncate-text;
    padding-right: var(--padding);
  }
}
</style>
