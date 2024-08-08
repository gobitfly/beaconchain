<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faInfoCircle,
} from '@fortawesome/pro-regular-svg-icons'

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const { setting, changeSetting } = useGlobalSetting<boolean>('rpl')

const rpActive = {
  get value(): boolean {
    return setting.value ?? true
  },
  set value(newValue: boolean) {
    changeSetting(newValue)
  },
}
</script>

<template>
  <div
    class="rp-row"
    :class="{ 'disable-in-production': !showInDevelopment }"
    @click.stop=""
  >
    <IconRocketPool class="icon" />
    <span class="text">
      {{ $t(`rocketpool.mode`) }}
    </span>
    <BcTooltip
      class="link"
      :text="$t('rocketpool.tooltip')"
    >
      <FontAwesomeIcon
        :icon="faInfoCircle"
        class="tooltip-icon"
      />
    </BcTooltip>
    <BcToggle
      v-model="rpActive.value"
      class="toggle"
      :disabled="!showInDevelopment"
    />
  </div>
</template>

<style lang="scss" scoped>
.rp-row {
  display: flex;
  align-items: center;

  &.disable-in-production {
    opacity: 0.5;
  }

  .link {
    z-index: 2;
  }

  .toggle {
    flex-grow: 1;
    justify-content: flex-end;
  }
}
</style>
