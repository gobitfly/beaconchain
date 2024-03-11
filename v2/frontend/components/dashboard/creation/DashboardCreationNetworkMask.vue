<script lang="ts" setup>
import { IconNetworkEthereumMono, IconNetworkGnosisMono } from '#components'

const { t: $t } = useI18n()

const network = defineModel<string>('network', { required: true })
const allNetworks = shallowRef([{ text: 'Ethereum', value: 'ethereum', component: IconNetworkEthereumMono }, { text: 'Gnosis', value: 'gnosis', component: IconNetworkGnosisMono }])

const emit = defineEmits<{(e: 'next'): void, (e: 'back'): void }>()

const continueDisabled = computed(() => {
  return network.value === ''
})
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
  .mask_container{
    .element_container{
      display: flex;
      flex-direction: column;
      gap: var(--padding);

      .row_container{
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
