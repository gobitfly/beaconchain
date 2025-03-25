<script setup lang="ts">
import type { Icon } from '~/components/bc/icon/BcIcon.vue'
import type {
  TagColor, TagSize,
} from '~/types/tag'

interface Props {
  color?: TagColor,
  icon?: Icon,
  label?: string,
  size?: TagSize,
  tooltip?: string,
}
defineProps<Props>()
</script>

<template>
  <BcTooltip
    v-if="label || icon"
    class="tag"
    :class="[color, size]"
    :text="tooltip"
    :fit-content="true"
  >
    {{ label }}
    <BcIcon
      v-if="icon"
      :name="icon"
    />
  </BcTooltip>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";
@use "~/assets/css/fonts.scss";

.tag {
  min-width: 80px;
  height: 20px;
  padding: 0 14px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 10px;
  @include fonts.tiny_text_bold;
  cursor: default;

  svg {
    margin-left: var(--padding-small);
    width: 12px;
    height: 12px;
  }

  &.compact {
    min-width: unset;
  }

  &.circle {
    min-width: unset;
    width: 14px;
    height: 14px;
    padding: 0;
    border-radius: 50%;
    font-size: var(--tooltip_text_font_size);

    svg {
      margin-left: unset;
      width: 10px;
      height: 10px;
    }
  }

  &.success {
    @include utils.positive-background;
  }

  &.failed {
    @include utils.negative-background;
  }

  &.orphaned {
    @include utils.orphaned-background;
  }

  &.partial {
    @include utils.partial-background;
  }

  &.light {
    @include utils.light-tag-background;
  }

  &.dark {
    @include utils.dark-tag-background;
  }
}
</style>
