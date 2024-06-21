<script lang="ts" setup>
const { t: $t } = useI18n()
const { currentPremiumSubscription } = useProductsStore()

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

// TODO: Only show a sinlge button based on "user is premium"
// TODO: Implement button logic and set buttonsDisabled accordingly

</script>

<template>
  <div class="subscriptions-container">
    <div class="title">
      {{ $t('user_settings.subscriptions.title') }}
    </div>
    <div class="subtitle">
      {{ $t('premium.title') }} | {{ currentPremiumSubscription?.product_name }}
    </div>
    <div class="explanation">
      {{ $t('user_settings.subscriptions.explanation') }}
    </div>
    <div class="button-row">
      <div class="manage-button">
        {{ $t('pricing.premium_product.button.manage_plan') }}
      </div>
      <Button :label="$t('pricing.premium_product.button.upgrade')" :disabled="buttonsDisabled" />
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

  .subtitle {
    @include fonts.subtitle_text;
  }

  .explanation {
    @include fonts.small_text;
    margin-bottom: var(--padding);
  }

  .button-row {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 30px;

    .manage-button {
      @include fonts.button_text;
      cursor: pointer;
    }
  }
}
</style>
