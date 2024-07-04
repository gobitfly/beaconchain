<script lang="ts" setup>
import { IconNetwork } from '#components'
import { ChainInfo, ChainID } from '~/types/network'
import { useNetworkStore } from '~/stores/useNetworkStore'

const { currentNetwork, availableNetworks, isNetworkDisabled } = useNetworkStore()

const network = defineModel<ChainID>('network')
const selection = ref<string>('')

watch(selection, (value) => { network.value = Number(value) as ChainID })

const buttonList = availableNetworks.value.map((chainId) => {
  return {
    value: String(chainId),
    className: 'dashboard-creation-singletoggle-button',
    text: ChainInfo[chainId].family as string,
    subText: (ChainInfo[chainId].name !== ChainInfo[chainId].family as string) ? ChainInfo[chainId].name : ChainInfo[chainId].description,
    disabled: isNetworkDisabled(chainId),
    component: IconNetwork,
    componentProps: { chainId, colored: false, harmonizePerceivedSize: true },
    componentClass: 'dashboard-creation-button-network-icon'
  }
})

const { t: $t } = useI18n()

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
        layout="gaudy"
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
  .dashboard-creation-singletoggle-button {
    width: 100%;
  }
  .dashboard-creation-button-network-icon {
    width: 100%;
    height: 100%;
  }
</style>
