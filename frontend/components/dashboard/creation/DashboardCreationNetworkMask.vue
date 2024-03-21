<script lang="ts" setup>
import { IconNetworkEthereum, IconNetworkGnosis } from '#components'

const { t: $t } = useI18n()

const network = defineModel<string>('network', { required: true })
const allNetworks = [{ text: 'Ethereum', value: 'ethereum', component: IconNetworkEthereum }, { text: 'Gnosis', value: 'gnosis', component: IconNetworkGnosis }]

const emit = defineEmits<{(e: 'next'): void, (e: 'back'): void }>()

const continueDisabled = computed(() => {
  return network.value === ''
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
      <BcToggleSingleBar v-model="network" class="single-bar" :buttons="allNetworks" :initial="network" />
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
