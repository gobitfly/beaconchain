<script lang="ts" setup>
const { t: $t } = useTranslation()

const notificationsManagementStore = useNotificationsManagementStore()

const networks = computed(() => notificationsManagementStore.settings.networks ?? [])
const currentNetworkId = computed(() => notificationsManagementStore.settings.networks[0]?.chain_id ?? 0)

const currentNetwork = computed(
  () => (networks.value.find(network => network.chain_id === currentNetworkId.value)),
)
const currentNetworkSettings = computed(() => currentNetwork.value?.settings)

const thresholdGasAbove = ref(formatWeiTo(currentNetworkSettings.value?.gas_above_threshold ?? '0', { unit: 'gwei' }))
const thresholdGasBelow = ref(formatWeiTo(currentNetworkSettings.value?.gas_below_threshold ?? '0', { unit: 'gwei' }))
const thresholdParticipationRate = ref(formatFraction(currentNetworkSettings.value?.participation_rate_threshold ?? 0))
const hasGasAbove = ref(currentNetworkSettings.value?.is_gas_above_subscribed ?? false)
const hasGasBelow = ref(currentNetworkSettings.value?.is_gas_below_subscribed ?? false)
const hasParticipationRate = ref(currentNetworkSettings.value?.is_participation_rate_subscribed ?? false)

watchDebounced([
  hasGasAbove,
  hasGasBelow,
  hasParticipationRate,
  thresholdGasAbove,
  thresholdGasBelow,
  thresholdParticipationRate,
], async () => {
  if (!currentNetworkSettings.value) return
  currentNetworkSettings.value.is_gas_above_subscribed = hasGasAbove.value
  currentNetworkSettings.value.gas_above_threshold = thresholdGasAbove.value
  currentNetworkSettings.value.is_gas_below_subscribed = hasGasBelow.value

  currentNetworkSettings.value.gas_above_threshold = formatToWei(thresholdGasAbove.value, { from: 'gwei' })
  currentNetworkSettings.value.gas_below_threshold = formatToWei(thresholdGasBelow.value, { from: 'gwei' })
  currentNetworkSettings.value.participation_rate_threshold = Number(formatToFraction(thresholdParticipationRate.value))

  await notificationsManagementStore.setNotificationForNetwork({
    chain_id: `${currentNetworkId.value}`,
    settings: currentNetworkSettings.value,
  })
},
{ deep: true },
)
</script>

<template>
  <div>
    <div v-if="!networks.length" class="error-container">
      Upps something went wrong. Please try again.
    </div>
    <BcTabPanel v-else>
      <!-- NetworkSwitcher:
        <button
          v-for="network in networks"
          :key="network.chain_id"
          :style="{ color: currentNetworkId === network.chain_id ? 'red' : 'black' }"
          @click="currentNetworkId = network.chain_id"
        >
          {{ network.chain_id }}
        </button> -->

      <div class="notifications-management-machines__content">
        <BcListSection class="grid-overwrite">
          <!-- <span class="grid-span-2">
              {{ $t('notifications.network.settings.new_reward_round') }}
            </span>
            <BcToggle
              v-model="hasNewRewardRound"
              class="toggle"
            /> -->
          <span>
            {{ $t('notifications.network.settings.alert_if_gas_below') }}
          </span>
          <span class="">
            <BcInputUnit
              v-model="thresholdGasBelow"
              :unit="$t('common.units.GWEI')"
            />
          </span>
          <BcToggle
            v-model="hasGasBelow"
            class="toggle"
          />
          <span>
            {{ $t('notifications.network.settings.alert_if_gas_above') }}
          </span>
          <span class="">
            <BcInputUnit
              v-model="thresholdGasAbove"
              :unit="$t('common.units.GWEI')"
            />
          </span>
          <BcToggle
            v-model="hasGasAbove"
            class="toggle"
          />
        </BcListSection>
        <BcListSection
          class="grid-overwrite"
          has-border-top
        >
          <span>
            {{ $t('notifications.network.settings.alert_if_participation_rate_below') }}
          </span>
          <span class="unit-overwrite">
            <BcInputUnit
              v-model="thresholdParticipationRate"
              unit="%"
            />
          </span>
          <BcToggle
            v-model="hasParticipationRate"
            class="toggle"
          />
        </BcListSection>
      </div>
    </BcTabPanel>
  </div>
</template>

<style lang="scss" scoped>
.error-container {
  text-align: center;
}
.unit-overwrite :deep(.bc-input-unit){
  margin-inline-end: 24px;
}
.grid-overwrite {
  grid-template-columns: 1fr auto auto;
}
.info{
  padding: var(--padding);
  text-align: center;
  color: var(--grey);
}
.is-text-disabled {
  color: var(--grey);
}

.toggle {
  justify-content: end;
  margin: 0;
}
.grid-span-2{
  grid-column: span 2;
}
.tutorial {
  text-align: center;
  padding: var(--padding);
}
.images {
  display: flex;
  gap: 10px;
  justify-content: center;
}
</style>
