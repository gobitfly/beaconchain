<script lang="ts" setup>
import { IconNetwork } from '#components'
import { ChainInfo, ChainIDs, isL1 } from '~/types/network'
import { useNetworkStore } from '~/stores/useNetworkStore'

const { currentNetwork, availableNetworks, isNetworkDisabled } = useNetworkStore()

const network = defineModel<ChainIDs>('network')
const selection = ref<string>('')

watch(selection, (value) => { network.value = Number(value) as ChainIDs })

const buttonList = computed(() => {
  const list = [] as any[]
  availableNetworks.value.forEach((chainId) => {
    if (isL1(chainId)) {
      list.push({
        value: String(chainId),
        text: ChainInfo[chainId].nameParts[0],
        subText: ChainInfo[chainId].nameParts[1],
        disabled: isNetworkDisabled(chainId),
        component: IconNetwork,
        componentProps: { chainId, colored: false, harmonizePerceivedSize: true },
        componentClass: 'dashboard-creation-button-network-icon'
      })
    }
  })
  return list
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
  .dashboard-creation-button-network-icon {
    width: 100%;
    height: 100%;
  }
</style>
