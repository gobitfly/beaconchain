<script lang="ts" setup>
import { faGem } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

defineProps<{
  screenreaderText?: string,
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
      to="/pricing"
      target="_blank"
      class="link"
    >
      <div>
        <LazyBcScreenreaderOnly v-if="screenreaderText">
          {{ screenreaderText }}
        </LazyBcScreenreaderOnly>
        <FontAwesomeIcon
          :icon="faGem"
          class="gem"
        />
      </div>
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
