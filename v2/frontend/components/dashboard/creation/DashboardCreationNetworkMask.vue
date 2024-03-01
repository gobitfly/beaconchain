<script lang="ts" setup>
const { t: $t } = useI18n()
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
      <div class="network_button_container">
        <Button>Ethereum</Button>
        <Button>Gnosis</Button>
      </div>
      <div class="continue_button_container">
        <Button>{{ $t('dashboard.creation.continue') }}</Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
  .mask_container{
    width: 460px;
    height: 248px;

    .element_container{
      display: flex;
      flex-direction: column;
      gap: 10px;

      .network_button_container{
        display: flex;
        gap: 10px;
      }

      .continue_button_container{
        display: flex;
        justify-content: flex-end;
      }
    }
  }
</style>
