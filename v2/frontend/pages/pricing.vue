<script lang="ts" setup>
import { faArrowDown } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const { t: $t } = useI18n()

useBcSeo('pricing.seo_title')

const { getProducts } = useProductsStore()
await useAsyncData('get_products', () => getProducts())

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
        <PricingPremiumProducts :is-yearly="isYearly" />
        <Button class="view-addons-button" @click="scrollToAddons()">
          {{ $t('pricing.view_addons') }}<FontAwesomeIcon :icon="faArrowDown" />
        </Button>
        <PricingPremiumCompare />
        <PricingPremiumAddons id="addons" :is-yearly="isYearly" />
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.page-container {
  width: 100%;

  display: flex;
  flex-direction: column;
  align-items: center;

  .page-content {
    // The pricing page uses unique styling, dimensions, font settings and so on that are not used anywhere else
    // That's why this component includes a lot of css
    // If a new page is introduced that uses the same parameters, consider moving them to a shared location
    font-family: var(--montserrat-family);
    font-weight: var(--montserrat-medium);

    max-width: var(--pricing-content-width-desktop);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding-top: 25px;

    .view-addons-button {
      width: 215px;
      height: 45px;
      font-size: 18px;
      display: flex;
      gap: 12px;
      margin-bottom: 35px;
    }
  }

  @media (max-width: 1360px) {
    .page-content {
      width: 100%;
      max-width: var(--pricing-content-width-mobile);

      .view-addons-button {
        padding: 7px 17px;
        width: 150px;
        height: 30px;
        font-size: 15px;
        gap: 8px;
      }
    }
  }
}
</style>
