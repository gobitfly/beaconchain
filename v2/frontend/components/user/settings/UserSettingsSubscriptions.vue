<script lang="ts" setup>
const { t: $t } = useI18n()
const { currentPremiumSubscription } = useProductsStore()
const { stripeCustomerPortal, isStripeDisabled } = useStripe()

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

async function goToStripeCustomerPortal () {
  if (planButton.value.disabled) {
    return
  }

  buttonsDisabled.value = true
  await stripeCustomerPortal()
  buttonsDisabled.value = false
}

const planButton = computed(() => {
  const text = currentPremiumSubscription.value ? $t('pricing.premium_product.button.manage_plan') : $t('pricing.premium_product.button.select_plan')
  const disabled = isStripeDisabled.value || buttonsDisabled.value || undefined

  return { text, disabled }
})

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
      <Button :label="planButton.text" :disabled="planButton.disabled" @click="goToStripeCustomerPortal()" />
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
  }
}
</style>
