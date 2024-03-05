<script lang="ts" setup>
import { IconMegaMenuEthereumOverview, IconMegaMenuGnosisOverview } from '#components'

const { t: $t } = useI18n()

const network = defineModel<string>({ required: true })
const allNetworks = ref([{ text: 'Ethereum', value: 'ethereum', component: IconMegaMenuEthereumOverview }, { text: 'Gnosis', value: 'gnosis', component: IconMegaMenuGnosisOverview }])
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
      <BcToggleSingleBar v-model="network" :buttons="allNetworks" />
      <div class="row_container">
        <Button>{{ $t('dashboard.creation.back') }}</Button>
        <Button>{{ $t('dashboard.creation.continue') }}</Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .mask_container{
    width: 460px;
    height: 248px;
    padding: var(--padding-large);

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
