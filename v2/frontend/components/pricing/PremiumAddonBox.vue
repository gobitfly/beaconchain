<script lang="ts" setup>
// TODOs
// - Check latest design changes (per month versus /month, asterisk, ...)
// - Implement changes for yearly/monthly view
// - Implement mobile

import { type ExtraDashboardValidatorsPremiumAddon } from '~/types/api/user'
import { formatPremiumProductPrice } from '~/utils/format'

const { t: $t } = useI18n()
const { products } = useProductsStore()

interface Props {
  addon: ExtraDashboardValidatorsPremiumAddon,
  isYearly: boolean
}
const props = defineProps<Props>()

const prices = computed(() => {
  const mainPrice = props.isYearly ? props.addon.price_per_year_eur / 12 : props.addon.price_per_month_eur

  const savingAmount = props.addon.price_per_month_eur * 12 - props.addon.price_per_year_eur
  const savingDigits = savingAmount % 100 === 0 ? 0 : 2

  return {
    main: formatPremiumProductPrice($t, mainPrice),
    monthly: formatPremiumProductPrice($t, props.addon.price_per_month_eur),
    monthly_based_on_yearly: formatPremiumProductPrice($t, props.addon.price_per_year_eur / 12),
    yearly: formatPremiumProductPrice($t, props.addon.price_per_year_eur),
    saving: formatPremiumProductPrice($t, savingAmount, savingDigits),
    perValidator: formatPremiumProductPrice($t, mainPrice / props.addon.extra_dashboard_validators, 5)
  }
})

const text = computed(() => {
  return {
    validatorCount: $t('pricing.addons.validator_amount', { amount: formatNumber(props.addon.extra_dashboard_validators) }),
    perValidator: $t('pricing.addons.per_validator', { amount: prices.value.perValidator })
  }
})
</script>

<template>
  <div class="box-container">
    <div class="summary-container">
      <div class="validator-count">
        {{ text.validatorCount }}
        <div class="subtext">
          {{ $t('pricing.addons.per_dashboard') }} i
        </div>
        <div class="per-validator">
          {{ text.perValidator }}
        </div>
      </div>
    </div>
    <div class="price-container">
      <div class="price">
        <span>{{ prices.monthly }}</span><span class="month"> {{ $t('pricing.addons.per_month') }}</span>
        <div class="year">
          {{ $t('pricing.addons.amount_per_year', {amount: prices.yearly}) }}
        </div>
      </div>
      <div class="saving-info">
        <div>
          {{ $t('pricing.addons.savings', {amount: prices.saving}) }}
        </div>
        <div>
          i
        </div>
      </div>
      <div class="quantity-container">
        <div>
          {{ $t('pricing.addons.quantity') }}
        </div>
        <InputText class="input">
          1
        </InputText>
      </div>
      <Button label="Select Add-On" class="select-button" />
      <div class="footer">
        {{ $t('pricing.addons.requires_plan', {name: products?.premium_products[products?.premium_products.length - 1].product_name}) }}
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  width: 348px;
  height: 100%;
  background-color: var(--container-background);
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  flex-shrink: 0;
  text-align: center;

  .summary-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    border-bottom: 2px solid var(--container-border-color);
    padding: 35px 0 26px 0;

    .validator-count {
      font-size: 24px;
      font-weight: 600;

      .subtext {
        font-weight: 400;
        margin-bottom: 16px;
      }

      .per-validator {
        color: var(--text-color-discreet);
        font-size: 20px;
        font-weight: 400;
      }
    }
  }

  .price-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 34px 9px 34px;

    .price {
      font-size: 32px;
      font-weight: 600;
      margin-bottom: 28px;

      .month {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 500;
      }

      .year {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 500;
      }
    }

    .saving-info {
      width: 100%;
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;
      gap: 13px;
      height: 37px;
      border-radius: 18px;
      background: var(--subcontainer-background);
      font-size: 17px;
      margin-bottom: 29px;
    }

    .quantity-container {
      display: flex;
      align-items: center;
      gap: 13px;
      font-size: 20px;
      margin-bottom: 32px;

      .input {
        width: 52px;
        border-radius: 9px;
      }
    }

    .select-button {
      width: 100%;
      height: 52px;
      font-size: 25px;
      font-weight: 500;
      margin-bottom: 26px;
    }

    .footer {
      width: 100%;
      text-align: right;
      font-size: 17px;
      font-weight: 400;
      color: var(--text-color-discreet);
    }
  }
}
</style>
