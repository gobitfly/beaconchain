<script lang="ts" setup>
const { t: $t } = useI18n()
const { products } = useProductsStore()

interface Props {
  isYearly: boolean
}
defineProps<Props>()
</script>

<template>
  <div class="premium-products-container">
    <div class="premium-products-row">
      <template v-for="product in products?.premium_products" :key="product.product_id">
        <PricingPremiumProductBox
          v-if="product.price_per_year_eur > 0"
          :product
          :is-yearly="isYearly"
        />
      </template>
    </div>
    <div class="footnote">
      {{ $t('pricing.excluding_vat') }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
.premium-products-container {
  width: 100%;

  .premium-products-row {
    display: flex;
    gap: 17px;
    justify-content: space-between;
    overflow-x: auto;
    padding-bottom: 7px;
  }

  .footnote {
    font-family: var(--roboto-family);
    font-size: 14px;
    font-weight: 400;
    color: var(--text-color-discreet);
    display: flex;
    justify-content: flex-end;
  }

  margin-bottom: 38px;

  @media (max-width: 600px) {
    .premium-products-container{
      margin-bottom: 36px;

      .footnote {
        font-size: 8px;
      }
    }
  }
}
</style>
