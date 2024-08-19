<script lang="ts" setup>
import { faGem } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
</script>

<template>
  <BcTooltip :text="$t('premium.subscribe')">
    <BcLink
      :to="`/pricing`"
      target="_blank"
      class="link"
    >
      <div>
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
