<script lang="ts" setup>
import { type DashboardCreationState } from '~/types/dashboard/creation'
import { IconNetworkEthereumMono, IconNetworkGnosisMono } from '#components'

const { t: $t } = useI18n()

const network = defineModel<string>('network', { required: true })
const state = defineModel<DashboardCreationState>('state', { required: true })
const allNetworks = shallowRef([{ text: 'Ethereum', value: 'ethereum', component: IconNetworkEthereumMono }, { text: 'Gnosis', value: 'gnosis', component: IconNetworkGnosisMono }])

const continueDisabled = computed(() => {
  return network.value === ''
})

function onContinue () {
  state.value = ''
}

function onBack () {
  state.value = 'type'
}
</script>

<template>
  <div class="mask_container">
    <div class="element_container">
      <div class="big_text">
        {{ $t('dashboard.creation.title') }}
      </div>
      <div class="subtitle_text">
        {{ $t('dashboard.creation.network.subtitle') }}
      </div>
      <BcToggleSingleBar v-model="network" :buttons="allNetworks" :initial="network" />
      <div class="row_container">
        <Button @click="onBack()">
          {{ $t('dashboard.creation.back') }}
        </Button>
        <Button :disabled="continueDisabled" @click="onContinue()">
          {{ $t('dashboard.creation.continue') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .mask_container{
    .element_container{
      display: flex;
      flex-direction: column;
      gap: 10px;

      .row_container{
        display: flex;
        justify-content: flex-end;
        gap: 10px;
      }

      button {
        width: 90px;
      }
    }
  }
</style>
