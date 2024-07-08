<script lang="ts" setup>
import { IconNetwork } from '#components'
import { ChainInfo, ChainIDs } from '~/types/network'
import { useNetworkStore } from '~/stores/useNetworkStore'

const { t: $t } = useI18n()

const { currentNetwork, isMainNet } = useNetworkStore()

const network = defineModel<ChainIDs>('network')
const selection = ref<`${ChainIDs}` | ''>('')
// TODO: get the list from the API...
let ValidatorDashboardNetworkList: ChainIDs[]
if (isMainNet()) {
  ValidatorDashboardNetworkList = [ChainIDs.Ethereum, ChainIDs.Gnosis]
  selection.value = `${ChainIDs.Ethereum}` // preselecting here, because Ethereum (Mainnet) is the only network that is available at the moment
} else {
  ValidatorDashboardNetworkList = [ChainIDs.Holesky, ChainIDs.Chiado]
  selection.value = `${ChainIDs.Holesky}` // preselecting here, because Ethereum (Holesky) is the only network that is available at the moment
}
// ... and remove this.

watch(selection, (value) => { network.value = Number(value) as ChainIDs }, { immediate: true })

const showNameOrDescription = (chainId: ChainIDs): string => {
  const chain = ChainInfo[chainId]
  if (chain.name === chain.family) {
    return chain.description
  }
  return chain.name
}

const buttonList = ValidatorDashboardNetworkList.map((chainId) => {
  // TODO: simply set `false` for everything once dashboards can be created for all the networks in `ValidatorDashboardNetworkList`
  const isDisabled = !useRuntimeConfig().public.showInDevelopment && chainId !== currentNetwork.value
  return {
    value: String(chainId),
    text: ChainInfo[chainId].family as string,
    subText: isDisabled ? $t('common.coming_soon') : showNameOrDescription(chainId),
    disabled: isDisabled,
    component: IconNetwork,
    componentProps: { chainId, colored: false, harmonizePerceivedSize: true },
    componentClass: 'dashboard-creation-button-network-icon'
  }
})

const emit = defineEmits<{(e: 'next'): void, (e: 'back'): void }>()

const continueDisabled = computed(() => {
  return !selection.value
})
</script>

<template>
  <div class="mask-container">
    <div class="element-container">
      <div class="big_text">
        {{ $t('dashboard.creation.title') }}
      </div>
      <div class="subtitle_text">
        {{ $t('dashboard.creation.network.subtitle') }}
      </div>
      <BcToggleSingleBar
        v-model="selection"
        class="single-bar"
        :buttons="buttonList"
        :initial="String(currentNetwork)"
        :are-buttons-networks="true"
      />
      <div class="row-container">
        <Button @click="emit('back')">
          {{ $t('navigation.back') }}
        </Button>
        <Button :disabled="continueDisabled" @click="emit('next')">
          {{ $t('navigation.continue') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .mask-container{
    width: 100%;
    .element-container{
      display: flex;
      flex-direction: column;
      gap: var(--padding);

      .single-bar{
        height: 100px;
      }

      .row-container{
        display: flex;
        justify-content: flex-end;
        gap: var(--padding);
      }

      button {
        width: 90px;
      }
    }
  }
</style>

<style lang="scss">
  .dashboard-creation-button-network-icon {
    width: 100%;
    height: 100%;
  }
</style>
