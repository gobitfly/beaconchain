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
  max-width: fit-content;

  .premium-products-row {
    display: flex;
    justify-content: space-between;
    overflow-x: auto;
    gap: 14px;
    padding-bottom: 4px;
  }

  .footnote {
    font-family: var(--roboto-family);
    font-size: 12px;
    font-weight: 400;
    color: var(--text-color-discreet);
    display: flex;
    justify-content: flex-end;
  }

  margin-bottom: 38px;
}

@media (max-width: 1360px) {
  .premium-products-container{
    max-width: fit-content;

    .premium-products-row {
      gap: 10px;
    }

    .footnote {
      font-size: 8px;
    }

    margin-bottom: 36px;
  }
}
</style>
