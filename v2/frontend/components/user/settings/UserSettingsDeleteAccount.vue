<script lang="ts" setup>
const { t: $t } = useI18n()

</script>

<template>
  <div class="subscriptions-container">
    <div class="title">
      {{ $t('user_settings.delete_account.title') }}
    </div>
    <div class="warning-row">
      <div class="warning">
        {{ $t('user_settings.delete_account.warning') }}
      </div>
      <Button :label="$t('user_settings.delete_account.button')" class="delete-button" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.subscriptions-container {
  display: flex;
  flex-direction: column;
  gap: var(--padding);

  padding: var(--padding-large);
  @include main.container;

  .title {
    @include fonts.dialog_header;
    margin-bottom: 9px;
  }

  .warning-row {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .warning {
      @include fonts.subtitle_text;
    }

    .delete-button {
      color: var(--button-text-color-disabled);
      background-color: var(--flashy-red);
      border-color: var(--flashy-red);
    }
  }
}
</style>
