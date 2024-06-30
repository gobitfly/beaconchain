<script setup lang="ts">
import type { MultiBarItem } from '~/types/multiBar'
const { availableNetworks } = useNetworkStore()

const { t } = useI18n()

const selection = ref<string[]>(['1'])

const buttons: MultiBarItem[] = [
  {
    component: undefined,
    value: String(1),
    tooltip: t('slotViz.filter.proposal')
  }
]

</script>

<template>
  <BcToggleMultiBar v-model="selection" :icons="buttons" />
</template>

<style scoped lang="scss">

</style>
