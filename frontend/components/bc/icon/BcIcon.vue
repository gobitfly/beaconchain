<script setup lang="ts">
import { LazyBcRotate } from '#components'
import type { Rotation } from '~/components/BcRotate.vue'

export type Icon
  = | 'alarm-snooze'
    | 'arrow-down'
    | 'arrow-left-right'
    | 'arrow-up'
    | 'arrow-upright-from-square'
    | 'bars'
    | 'bolt'
    | 'chart-column'
    | 'chart-line-up'
    | 'check'
    | 'chevron-down'
    | 'chevron-left'
    | 'chevron-right'
    | 'chevrons-left'
    | 'circle-info'
    | 'circle-plus'
    | 'circle-user'
    | 'clock'
    | 'copy'
    | 'cube'
    | 'cubes'
    | 'desktop'
    | 'edit'
    | 'eye'
    | 'file-signature'
    | 'fire'
    | 'flag'
    | 'gauge-simple-max'
    | 'gear'
    | 'gem'
    | 'globe'
    | 'hand-holding-hand'
    | 'hashtag'
    | 'magnifying-glass'
    | 'minus'
    | 'money-bill'
    | 'monitor-wave-form'
    | 'moon'
    | 'network-wired'
    | 'paper-plane'
    | 'people-group'
    | 'percent'
    | 'plus'
    | 'power-off'
    | 'right-from-bracket'
    | 'rocket'
    | 'rotate'
    | 'share'
    | 'sigma'
    | 'snooze'
    | 'sun'
    | 'sync'
    | 'table'
    | 'trash'
    | 'user'
    | 'user-slash'
    | 'wallet'
    | 'xmark'
    | IconCrypto
    | IconCustom
    | IconFlag

type IconCrypto
  = | 'crypto-eth'
    | 'crypto-gno'
    | 'crypto-xdai'

type IconCustom
  = | 'attestaion-head'
    | 'attestaion-source'
    | 'attestaion-target'
    | 'circle-power-off'

type IconFlag
  = | 'flag-au'
    | 'flag-ca'
    | 'flag-cn'
    | 'flag-eu'
    | 'flag-gb'
    | 'flag-jp'
    | 'flag-us'

const {
  rotation,
  size = 'base',
} = defineProps<{
  name: Icon,
  rotation?: Rotation,
  size?: 'base' | 'sm',
}>()

const is = computed(() => {
  if (rotation) return LazyBcRotate
  return 'span'
})
</script>

<template>
  <component
    :is
    :rotation
    class="bc-icon"
    :class="{
      base: size === 'base',
      sm: size === 'sm',
    }"
  >
    <LazyBcIconAlarmSnooze v-if="name === 'alarm-snooze'" />
    <LazyBcIconArrowDown v-else-if="name === 'arrow-down'" />
    <LazyBcIconArrowLeftRight v-else-if="name === 'arrow-left-right'" />
    <LazyBcIconArrowUp v-else-if="name === 'arrow-up'" />
    <LazyBcIconArrowUpRightFromSquare v-else-if="name === 'arrow-upright-from-square'" />
    <LazyBcIconBars v-else-if="name === 'bars'" />
    <LazyBcIconBolt v-else-if="name === 'bolt'" />
    <LazyBcIconChartColumn v-else-if="name === 'chart-column'" />
    <LazyBcIconChartLineUp v-else-if="name === 'chart-line-up'" />
    <LazyBcIconCheck v-else-if="name === 'check'" />
    <LazyBcIconChevronDown v-else-if="name === 'chevron-down'" />
    <LazyBcIconChevronLeft v-else-if="name === 'chevron-left'" />
    <LazyBcIconChevronRight v-else-if="name === 'chevron-right'" />
    <LazyBcIconChevronsLeft v-else-if="name === 'chevrons-left'" />
    <LazyBcIconCircleInfo v-else-if="name === 'circle-info'" />
    <LazyBcIconCirclePlus v-else-if="name === 'circle-plus'" />
    <LazyBcIconCircleUser v-else-if="name === 'circle-user'" />
    <LazyBcIconClock v-else-if="name === 'clock'" />
    <LazyBcIconCopy v-else-if="name === 'copy'" />
    <LazyBcIconCube v-else-if="name === 'cube'" />
    <LazyBcIconCubes v-else-if="name === 'cubes'" />
    <LazyBcIconDesktop v-else-if="name === 'desktop'" />
    <LazyBcIconEdit v-else-if="name === 'edit'" />
    <LazyBcIconEye v-else-if="name === 'eye'" />
    <LazyBcIconFileSignature v-else-if="name === 'file-signature'" />
    <LazyBcIconFire v-else-if="name === 'fire'" />
    <LazyBcIconFlag v-else-if="name === 'flag'" />
    <LazyBcIconGaugeSimpleMax v-else-if="name === 'gauge-simple-max'" />
    <LazyBcIconGear v-else-if="name === 'gear'" />
    <LazyBcIconGem v-else-if="name === 'gem'" />
    <LazyBcIconGlobe v-else-if="name === 'globe'" />
    <LazyBcIconHandHoldingHand v-else-if="name === 'hand-holding-hand'" />
    <LazyBcIconHashtag v-else-if="name === 'hashtag'" />
    <LazyBcIconMagnifyingGlass v-else-if="name === 'magnifying-glass'" />
    <LazyBcIconMinus v-else-if="name === 'minus'" />
    <LazyBcIconMoneyBill v-else-if="name === 'money-bill'" />
    <LazyBcIconMonitorWaveForm v-else-if="name === 'monitor-wave-form'" />
    <LazyBcIconMoon v-else-if="name === 'moon'" />
    <LazyBcIconNetworkWired v-else-if="name === 'network-wired'" />
    <LazyBcIconPaperPlane v-else-if="name === 'paper-plane'" />
    <LazyBcIconPeopleGroup v-else-if="name === 'people-group'" />
    <LazyBcIconPercent v-else-if="name === 'percent'" />
    <LazyBcIconPlus v-else-if="name === 'plus'" />
    <LazyBcIconPowerOff v-else-if="name === 'power-off'" />
    <LazyBcIconRightFromBracket v-else-if="name === 'right-from-bracket'" />
    <LazyBcIconRocket v-else-if="name === 'rocket'" />
    <LazyBcIconRotate v-else-if="name === 'rotate'" />
    <LazyBcIconShare v-else-if="name === 'share'" />
    <LazyBcIconSigma v-else-if="name === 'sigma'" />
    <LazyBcIconSnooze v-else-if="name === 'snooze'" />
    <LazyBcIconSun v-else-if="name === 'sun'" />
    <LazyBcIconSync v-else-if="name === 'sync'" />
    <LazyBcIconTable v-else-if="name === 'table'" />
    <LazyBcIconTrash v-else-if="name === 'trash'" />
    <LazyBcIconUser v-else-if="name === 'user'" />
    <LazyBcIconUserSlash v-else-if="name === 'user-slash'" />
    <LazyBcIconWallet v-else-if="name === 'wallet'" />
    <LazyBcIconXmark v-else-if="name === 'xmark'" />

    <LazyBcIconCryptoEth v-else-if="name === 'crypto-eth'" />
    <LazyBcIconCryptoGno v-else-if="name === 'crypto-gno'" />
    <LazyBcIconCryptoXdai v-else-if="name === 'crypto-xdai'" />

    <LazyBcIconAttestationHead v-else-if="name === 'attestaion-head'" />
    <LazyBcIconAttestationSource v-else-if="name === 'attestaion-source'" />
    <LazyBcIconAttestationTarget v-else-if="name === 'attestaion-target'" />
    <LazyBcIconCirclePowerOff v-else-if="name === 'circle-power-off'" />

    <LazyBcIconFlagAu v-else-if="name === 'flag-au'" />
    <LazyBcIconFlagCa v-else-if="name === 'flag-ca'" />
    <LazyBcIconFlagCn v-else-if="name === 'flag-cn'" />
    <LazyBcIconFlagEu v-else-if="name === 'flag-eu'" />
    <LazyBcIconFlagGb v-else-if="name === 'flag-gb'" />
    <LazyBcIconFlagJp v-else-if="name === 'flag-jp'" />
    <LazyBcIconFlagUs v-else-if="name === 'flag-us'" />
  </component>
</template>

<style scoped lang="scss">
.bc-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.base {
  :deep(svg) {
    width: 1rem;
  }
}
.sm {
  :deep(svg) {
    width: .75rem;
  }
}
</style>
