<script lang="ts" setup>
const { products } = useProductsStore()

const isYearly = defineModel<boolean>({ required: true })

const savingPercentage = computed(() => {
  let highestSaving = 0
  products.value?.premium_products.forEach((product) => {
    const savingPercentage = (1 - (product.price_per_year_eur / (product.price_per_month_eur * 12))) * 100
    if (savingPercentage > highestSaving) {
      highestSaving = savingPercentage
    }
  })

  return Math.floor(highestSaving)
})
</script>

<template>
  <div class="toggle-container">
    <BcToggle
      v-model="isYearly"
      class="toggle"
      :true-option="$t('pricing.yearly')"
      :false-option="$t('pricing.monthly')"
    />
    <div
      v-if="savingPercentage > 0"
      class="save-up-text"
    >
      {{ $t('pricing.save_up_to', { percentage: savingPercentage }) }}
    </div>
  </div>
</template>

<style lang="scss">
.toggle-container {
  display: flex;
  align-items: center;
  gap: var(--padding);
  margin-bottom: 55px;

  .toggle{
    font-size: 18px;
    font-weight: var(--montserrat-light);
    margin-bottom: 0;
  }

  .save-up-text {
    width: 75px;
    color: var(--primary-color);
    text-align: center;
    font-size: 13px;
    font-weight: var(--montserrat-semi-bold);
  }
}

@media (max-width: 1360px) {
  .toggle-container {
    font-size: 16px;
    margin-bottom: 30px;

    .toggle {
      font-size: 16px;
    }

    .save-up-text {
      width: 60px;
      font-size: 12px;
    }
  }
}
</style>
