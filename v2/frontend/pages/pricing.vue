<script lang="ts" setup>
import { faArrowDown } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const { t: $t } = useI18n()

useBcSeo('pricing.seo_title')
const { stripeInit } = useStripeProvider()

const { products, getProducts } = useProductsStore()

await useAsyncData('get_products', () => getProducts())
watch(products, () => {
  if (products.value?.stripe_public_key) {
    stripeInit(products.value.stripe_public_key)
  }
}, { immediate: true })

const isYearly = ref(true)

const scrollToAddons = () => {
  const element = document.getElementById('addons')
  element?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<template>
  <BcPageWrapper>
    <div class="page-container">
      <div class="page-content">
        <PricingTypeToggle />
        <PricingHeaderLine />
        <PricingPeriodToggle v-model="isYearly" />
        <PricingPremiumViaAppBanner />
        <PricingPremiumProducts :is-yearly="isYearly" />
        <Button class="view-addons-button" @click="scrollToAddons()">
          {{ $t('pricing.view_addons') }}<FontAwesomeIcon :icon="faArrowDown" />
        </Button>
        <PricingPremiumCompare />
        <PricingPremiumAddons id="addons" :is-yearly="isYearly" />
        <BcFaq class="faq" translation-path="faq.pricing" />
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss">
@import "~/assets/css/pricing.scss";
</style>

<style lang="scss" scoped>
@use '~/assets/css/pricing.scss';

.page-container {
  width: 100%;

  display: flex;
  flex-direction: column;
  align-items: center;

  .page-content {
    // The pricing page uses unique styling, dimensions, font settings and so on that are not used anywhere else
    // That's why this component and its children include a lot of handcraftet css
    // If a new page is introduced that uses the same parameters, consider moving them to a shared location
    font-family: var(--montserrat-family);
    font-weight: var(--montserrat-medium);

    max-width: var(--pricing-content-width);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding-top: 25px;

    .view-addons-button {
      width: 215px;
      @include pricing.pricing_button;
      display: flex;
      gap: 12px;
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
  .faq{
    width: 100%;
    margin-top: 51px;
  }
}
</style>
