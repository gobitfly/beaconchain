<script lang="ts" setup>
const { t: $t } = useI18n()
const { products, bestPremiumProduct } = useProductsStore()
const { user } = useUserStore()

interface Props {
  isYearly: boolean
}
defineProps<Props>()

const addonsAvailable = computed(() => {
  if (!products?.value?.premium_products || !bestPremiumProduct?.value) {
    return false
  }

  return user.value?.subscriptions.find(sub => sub.product_id === bestPremiumProduct.value?.product_id) !== undefined
})
</script>

<template>
  <div class="addons-container">
    <div class="text-container">
      <div class="title">
        {{ $t('pricing.addons.title') }}
      </div>
      <div class="subtitle">
        {{ $t('pricing.addons.subtitle') }}
      </div>
    </div>
    <div class="addons-row">
      <PricingPremiumAddonBox
        v-for="addon in products?.extra_dashboard_validators_premium_addons"
        :key="addon.product_id"
        :addon="addon"
        :addons-available="addonsAvailable"
        :is-yearly="isYearly"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.addons-container {
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 10px;

  .text-container {
    display: flex;
    flex-direction: column;

    .title {
      font-size: 26px;
      color: var(--primary-color);
    }

    .subtitle {
      font-size: 29px;
    }
  }

  .addons-row {
    width: 100%;
    max-width: fit-content;
    flex-shrink: 0;
    display: flex;
    justify-content: space-between;
    overflow-x: auto;
    gap: 7px;
    padding-bottom: 4px;
  }

  @media (max-width: 1360px) {
    flex-direction: column;
    gap: 15px;

    .text-container {
      gap: 8px;

      .title{
        font-size: 16px;
      }

      .subtitle{
        font-size: 18px;
      }
    }
  }
}
</style>
