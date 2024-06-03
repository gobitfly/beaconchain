<script lang="ts" setup>
import { ChainInfo, ChainIDs } from '~/types/networks'
import { useNetworkStore } from '~/stores/useNetworkStore'

// TODO: Get this list from the API
const ValidatorDashboardNetworkList = [ChainIDs.Ethereum, ChainIDs.Holesky, ChainIDs.Gnosis]

const availableNetworks = useNetworkStore().getAvailableNetworks()

const network = defineModel<ChainIDs>('network')
const selection = ref<string>('')

watch(selection, (value) => { network.value = Number(value) as ChainIDs })

const buttonList = ValidatorDashboardNetworkList.map((chainId) => {
  return {
    value: String(chainId),
    text: ChainInfo[chainId].name,
    disabled: !(chainId in availableNetworks)
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
        :initial="network ? String(network) : ''"
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
