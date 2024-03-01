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
        {{ $t('dashboard.creation.type.subtitle') }}
      </div>
      <div class="row_container">
        <Button>{{ $t('dashboard.creation.type.accounts') }}</Button>
        <Button>{{ $t('dashboard.creation.type.validators') }}</Button>
      </div>
      <div class="row_container">
        <InputText />
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

      .row_container{
        display: flex;
        gap: 10px;
      }
    }
  }
</style>
