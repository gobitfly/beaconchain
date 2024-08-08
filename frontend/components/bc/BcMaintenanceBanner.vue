<script setup lang="ts">
import { warn } from 'vue'

const {
  public: { maintenanceTS },
} = useRuntimeConfig()
const { tick } = useInterval(60)
const { t: $t } = useTranslation()

const maintenanceLabel = computed(() => {
  if (!maintenanceTS) {
    return
  }
  const parsed
    = typeof maintenanceTS === 'number' ? maintenanceTS : parseInt(maintenanceTS)
  if (isNaN(parsed)) {
    warn(
      'NUXT_PUBLIC_MAINTENANCE_TS is not convertible to an integer, a unix ts is expected',
    )
    return undefined
  }
  else if (parsed === 0) {
    return
  }
  const ts = new Date(parsed * 1000).getTime()
  if (ts > tick.value) {
    return $t('maintenance.planned', {
      date: formatTsToAbsolute(ts / 1000, $t('locales.date'), true),
    })
  }
  else {
    return $t('maintenance.ongoing')
  }
})
</script>

<template>
  <div
    v-if="maintenanceLabel"
    class="maintenance-banner"
  >
    {{ maintenanceLabel }}
  </div>
</template>

<style lang="scss" scoped>
.maintenance-banner {
  width: 100%;
  padding: var(--padding-large);
  background-color: var(--grey-4);
  //color: var(--container-color);
  text-align: center;
}
.dark-mode {
  .maintenance-banner {
    background-color: var(--very-dark-grey);
  }
}
</style>
