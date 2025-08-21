<script lang="ts" setup>
defineProps<{
  screenreaderText?: TranslationInput,
  toolTipText?: string,
  tooltipWidth?: `${number}%` | `${number}px`,
}>()
</script>

<template>
  <BcTooltip
    :tooltip-width
    :text="toolTipText ?? $t('premium.subscribe')"
  >
    <BcLink
      to="/premium"
      target="_blank"
      class="link"
    >
      <span>
        <LazyBcScreenreaderOnly
          v-if="screenreaderText"
          :screenreader-text
        />
        <BcIcon
          class="gem"
          name="gem"
        />
      </span>
    </BcLink>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.gem {
  color: var(--primary-color);
  cursor: pointer;
}

.link {
  cursor: pointer;
  pointer-events: all;
}
</style>
