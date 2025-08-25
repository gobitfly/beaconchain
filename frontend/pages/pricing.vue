<script lang="ts" setup>
const { t: $t } = useTranslation()

useBcSeo('pricing.seo_title')
const { promoCode } = usePromoCode()
const { stripeInit } = useStripeProvider()

const {
  getProducts,
  products,
} = useProductsStore()

await useAsyncData('get_products', () => getProducts())
watch(
  products,
  () => {
    if (products.value?.stripe_public_key) {
      stripeInit(products.value.stripe_public_key)
    }
  },
  { immediate: true },
)

const isPaymentYearly = ref(true)

const scrollToAddons = () => {
  const element = document.getElementById('addons')
  element?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<template>
  <div>
    <div class="page-container">
      <div class="page-content">
        <div class="type-toggle-row">
          <div class="content">
            <BcButtonIcon
              screenreader-text="navigation.back"
              name="arrow-down"
              class="back-button"
              @click="$router.back()"
            />
            <PricingTypeToggle />
          </div>
        </div>
        <PricingHeaderLine />
        <PricingPeriodToggle v-model="isPaymentYearly" />
        <PricingPremiumViaAppBanner />
        <PricingPremiumProducts :is-payment-yearly />
        <BcButton
          class="view-addons-button"
          @click="scrollToAddons()"
        >
          {{ $t("pricing.view_addons") }}
          <BcIcon name="arrow-down" />
        </BcButton>
        <PricingPremiumCompare />
        <PricingPremiumAddons
          id="addons"
          :is-payment-yearly
        />
      </div>
      <div
        v-if="promoCode"
        class="promo-overlay"
      >
        <I18nT
          keypath="pricing.promo_code"
          scope="global"
          tag="span"
          class="promo-text"
        >
          <template #_code>
            <span class="promo-code">{{ promoCode }}</span>
          </template>
        </I18nT>
      </div>
    </div>
  </div>
</template>

<style lang="css">
:root {
  --pricing-content-width: 800px;

  @media (min-width: 1360px) {
    --pricing-content-width: 1000px;
  }
}
</style>

<style lang="scss" scoped>
.promo-overlay {
  position: fixed;
  z-index: 6;
  bottom: 1px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  pointer-events: none;
  .promo-text {
    background-color: var(--background-color);
    border: 1px solid var(--primary-color);
    border-radius: 4px;
    padding: 18px 26px;
    pointer-events: unset;

    .promo-code {
      color: var(--primary-color);
    }
  }
}

.page-container {
  position: relative;
  width: 100%;

  display: flex;
  flex-direction: column;
  align-items: center;

  .page-content {
    // The pricing page uses unique styling, dimensions, font settings and so on that are not used anywhere else
    // That's why this component and its children include a lot of handcraftet css
    // If a new page is introduced that uses the same parameters, consider moving them to a shared location
    position: relative;
    font-family: var(--montserrat-family);
    font-weight: var(--montserrat-medium);

    max-width: var(--pricing-content-width);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    .type-toggle-row {
      display: flex;
      position: sticky;
      z-index: 128;
      top: 0px;
      padding-top: 25px;
      padding-bottom: 25px;
      margin-bottom: 12px;
      background-color: var(--background-color);
      width: 100vw;
      justify-content: center;
      .content {
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        padding-inline: var(--padding-small);
        width: var(--pricing-content-width);

        .back-button{
          rotate: 90deg;
          justify-self: flex-start;
        }
      }
    }

    .view-addons-button {
      margin-bottom: 35px;
    }
  }

  @media (max-width: 1360px) {
    .page-content {
      width: 100%;

      .view-addons-button {
        padding: 7px 17px;
        width: 150px;
        gap: 8px;
      }
    }
  }
}
</style>
