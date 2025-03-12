<script setup lang="ts">
import { faEdit } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const emit = defineEmits<{ (e: 'onEdit'): void }>()

defineProps<{
  isDisabled?: boolean,
  label?: string,
  truncateText?: boolean,
}>()

const { t: $t } = useTranslation()
</script>

<template>
  <div
    class="bc-poput-edit"
    :class="{ 'truncate-text': truncateText }"
  >
    <slot name="content">
      <BcTooltip
        v-if="label"
        tooltip-width="320px"
        tooltip-text-align="left"
        :hide="!truncateText"
        class="content"
        :text="label"
      >
        {{ label }}
      </BcTooltip>
    </slot>
    <div class="icon">
      <BcButtonIcon
        :screenreader-text="$t('common.edit')"
        :is-disabled
        class="link"
        @click="() => emit('onEdit')"
      >
        <FontAwesomeIcon
          :icon="faEdit"
        />
      </BcButtonIcon>
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

  .link {
    &:disabled {
      cursor: auto;
    }
  }
}
</style>
