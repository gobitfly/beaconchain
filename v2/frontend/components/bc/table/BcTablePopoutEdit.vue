<script setup lang="ts">
import { faEdit } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const emit = defineEmits<{ (e: 'onEdit'): void }>()

interface Props {
  label?: string,
  noIcon?: boolean,
  truncateText?: boolean,
}
defineProps<Props>()
</script>

<template>
  <div
    class="bc-poput-edit"
    :class="{ 'truncate-text': truncateText }"
  >
    <slot name="content">
      <BcTooltip
        v-if="label"
        :hide="!truncateText"
        :fit-content="true"
        class="content"
        :text="label"
      >
        {{ label }}
      </BcTooltip>
    </slot>
    <div class="icon">
      <FontAwesomeIcon
        v-if="!noIcon"
        class="link"
        :icon="faEdit"
        @click="() => emit('onEdit')"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.bc-poput-edit {
  display: flex;

  &.truncate-text {
    align-items: center;

    .content {
      @include utils.truncate-text;
      user-select: none;
    }
  }

  &:not(.truncate-text) {
    .icon {
      flex-grow: 1;
      display: flex;
      justify-content: flex-end;
    }
  }

  .content {
    padding-right: var(--padding);
  }
}
</style>
