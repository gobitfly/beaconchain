<script lang="ts" setup>
const { t } = useI18n()

const { products, getProducts } = useProductsStore()
await getProducts()

const isYearly = ref(true)

const savingPercentage = computed(() => {
  let highestSaving = 0
  products.value?.data.premium_products.forEach((product) => {
    const savingPercentage = (1 - (product.price_per_year_eur / (product.price_per_month_eur * 12))) * 100
    if (savingPercentage > highestSaving) {
      highestSaving = savingPercentage
    }
  })

  return Math.floor(highestSaving)
})

</script>

<template>
  <BcPageWrapper>
    <div class="page-container">
      <div class="type-toggle-container">
        <div class="premium">
          <div class="text">
            {{ t('pricing.premium') }}
          </div>
        </div>
        <div class="api-keys" disabled>
          <div class="text">
            {{ t('pricing.API_keys') }}
          </div>
        </div>
      </div>
      <div class="header-line-container">
        <div class="header-line">
          <div class="title">
            {{ t('pricing.premium') }}
          </div>
          <div class="subtitle">
            {{ t('pricing.subtitle') }}
          </div>
        </div>
      </div>
      <div class="toggle-container">
        <BcToggle
          v-model="isYearly"
          class="toggle"
          :true-option="t('pricing.yearly')"
          :false-option="t('pricing.monthly')"
        />
        <div v-if="savingPercentage > 0" class="save-up-text">
          {{ t('pricing.save_up_to', {percentage: savingPercentage}) }}
        </div>
      </div>
      <div class="premium-products-container">
        <div class="premium-products-row">
          <div v-for="product in products?.data.premium_products" :key="product.product_id">
            <PricingPremiumProductBox
              v-if="product.price_per_year_eur > 0"
              :product
              :is-yearly="isYearly"
            />
          </div>
        </div>
        <div class="footnote">
          {{ t('pricing.excluding_vat') }}
        </div>
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.page-container {
  // The pricing page uses unique styling, dimensions, font settings and so on that are not used anywhere else
  // That's why this component includes a lot of css
  // If a new page is introduced that uses the same parameters, consider moving them to a shared location
  font-family: var(--montserrat-family);
  font-weight: var(--montserrat-medium);

  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  .type-toggle-container {
    width: 237px;
    height: 41px;
    margin-top: 25px;
    margin-bottom: 37px;
    display: flex;
    align-items: center;
    user-select: none;

    // not interactive for beta launch
    .premium {
      flex: 1;
      height: 100%;
      display: flex;
      align-items: center;
      border-top-left-radius: 3px;
      border-bottom-left-radius: 3px;
      background: var(--button-color-active);

      .text {
        color: var(--primary-contrast-color);
        font-size: 13px;
        font-weight: var(--montserrat-semi-bold);
        padding-left: 16px;
      }
    }

    .api-keys {
      flex: 1;
      height: 100%;
      box-sizing: border-box;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--container-background);
      border-width: 1px 1px 1px 0;
      border-style: solid;
      border-color: var(--container-border-color);
      border-top-right-radius: 3px;
      border-bottom-right-radius: 3px;

      .text {
        color: var(--grey);
        font-size: 14px;
        font-weight: var(--montserrat-semi-bold);
      }
    }
  }

  .header-line-container {
      width: 100vw;
      display: flex;
      justify-content: center;
      padding: 21px 0;
      border: 1px solid var(--container-border-color);
      background: var(--container-background);
      margin-bottom: 55px;

      .header-line {
        width: var(--content-width);
        display: flex;
        flex-direction: column;

        .title {
          font-size: 18px;
          color: var(--primary-color);
          margin-bottom: var(--padding);
        }

        .subtitle {
          font-size: 32px;
        }
    }
  }

  .toggle-container {
    display: flex;
    align-items: center;
    gap: var(--padding);
    margin-bottom: 55px;

    .toggle{
      font-size: 20px;
      margin-bottom: 0;
    }

    .save-up-text {
      width: 75px;
      color: var(--primary-color);
      text-align: center;
      font-size: 15px;
      font-weight: var(--montserrat-semi-bold);
    }
  }

  .premium-products-container {
    height: min-content;

    .premium-products-row {
      display: flex;
      gap: 17px;
      justify-content: space-between;
      margin-bottom: 8px;
    }

    .footnote{
      font-size: 21px;
      font-weight: 400;
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>
