<script lang="ts" setup>
import {
  faGem
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
</script>

<template>
  <BcTooltip :text="$t('premium.subscribe')">
    <NuxtLink :to="`/pricing`" target="_blank" :no-prefetch="true">
      <div>
        <FontAwesomeIcon :icon="faGem" class="gem" />
      </div>
    </NuxtLink>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.gem {
  color: var(--primary-color);
  cursor: pointer;
}
</style>
