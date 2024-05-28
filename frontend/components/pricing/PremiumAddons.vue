<script lang="ts" setup>
const { t: $t } = useI18n()
const { products } = useProductsStore()
const { user } = useUserStore()

interface Props {
  isYearly: boolean
}
defineProps<Props>()

const addonsAvailable = computed(() => {
  return user.value?.subscriptions.find(sub => sub.product_id === products?.value?.premium_products[products?.value.premium_products.length - 1].product_id) !== undefined
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
      <template v-for="addon in products?.extra_dashboard_validators_premium_addons" :key="addon.product_id">
        <PricingPremiumAddonBox :addon="addon" :addons-available="addonsAvailable" :is-yearly="isYearly" />
      </template>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.addons-container {
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 70px;

  .text-container {
    display: flex;
    flex-direction: column;

    .title {
      font-size: 32px;
      color: var(--primary-color);
    }

    .subtitle {
      font-size: 35px;
    }
  }

  .addons-row {
    width: 100%;
    display: flex;
    gap: 7px;
    overflow-x: auto;
  }

  @media (max-width: 600px) {
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
