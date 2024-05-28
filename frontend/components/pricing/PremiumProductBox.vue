<script lang="ts" setup>
// TODO: Add links to Buttons (don't forget Downgrade "button")

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumProduct } from '~/types/api/user'
import { formatPremiumProductPrice } from '~/utils/format'
import type { Feature } from '~/types/pricing'
// import { formatTimeDuration } from '~/utils/format' TODO: See commented code below

const { products } = useProductsStore()
const { user } = useUserStore()
const { t: $t } = useI18n()

interface Props {
  product: PremiumProduct,
  isYearly: boolean
}
const props = defineProps<Props>()

const prices = computed(() => {
  const mainPrice = props.isYearly ? props.product.price_per_year_eur / 12 : props.product.price_per_month_eur

  const savingAmount = props.product.price_per_month_eur * 12 - props.product.price_per_year_eur
  const savingDigits = savingAmount % 100 === 0 ? 0 : 2

  return {
    main: formatPremiumProductPrice($t, mainPrice),
    monthly: formatPremiumProductPrice($t, props.product.price_per_month_eur),
    monthly_based_on_yearly: formatPremiumProductPrice($t, props.product.price_per_year_eur / 12),
    yearly: formatPremiumProductPrice($t, props.product.price_per_year_eur),
    saving: formatPremiumProductPrice($t, savingAmount, savingDigits),
    perValidator: formatPremiumProductPrice($t, mainPrice / props.product.premium_perks.validators_per_dashboard, 5)
  }
})

const percentages = computed(() => {
  // compare with the last product in the list
  const compareProduct = products.value?.premium_products[products.value.premium_products.length - 1]

  return {
    validatorDashboards: props.product.premium_perks.validator_dashboards / (compareProduct?.premium_perks.validator_dashboards ?? 1) * 100,
    validatorsPerDashboard: props.product.premium_perks.validators_per_dashboard / (compareProduct?.premium_perks.validators_per_dashboard ?? 1) * 100,
    summaryChart: props.product.premium_perks.summary_chart_history_seconds / (compareProduct?.premium_perks.summary_chart_history_seconds ?? 1) * 100,
    heatmapChart: props.product.premium_perks.heatmap_history_seconds / (compareProduct?.premium_perks.heatmap_history_seconds ?? 1) * 100
  }
})

const planButton = computed(() => {
  let isDowngrade = false
  let text = $t('pricing.premium_product.button.select_plan')

  if (user.value?.subscriptions) {
    const subscription = user.value?.subscriptions?.find(sub => sub.product_category === 'premium')
    if (!subscription) {
      text = $t('pricing.premium_product.button.select_plan')
    } else if (subscription.product_id === props.product.product_id) {
      text = $t('pricing.premium_product.button.manage_plan')
    } else if (subscription.product_id < props.product.product_id) {
      text = $t('pricing.premium_product.button.upgrade')
    } else {
      isDowngrade = true
      text = $t('pricing.premium_product.button.downgrade')
    }
  }

  return { text, isDowngrade }
})

const mainFeatures = computed<Feature[]>(() => {
  return [
    {
      name: $t('pricing.premium_product.validator_dashboards', { amount: formatNumber(props.product?.premium_perks.validator_dashboards) }, (props.product?.premium_perks.validator_dashboards || 0) <= 1 ? 1 : 2),
      available: true,
      percentage: percentages.value.validatorDashboards
    },
    {
      name: $t('pricing.premium_product.validators_per_dashboard', { amount: formatNumber(props.product?.premium_perks.validators_per_dashboard) }),
      subtext: $t('pricing.per_validator', { amount: prices.value.perValidator }),
      available: true,
      tooltip: $t('pricing.pectra_tooltip', { effectiveBalance: formatNumber(props.product?.premium_perks.validators_per_dashboard * 32) }),
      percentage: percentages.value.validatorsPerDashboard
    },
    {
      name: $t('pricing.premium_product.timeframe_dashboard_chart_no_timeframe'),
      subtext: $t('pricing.premium_product.coming_soon'),
      available: true,
      percentage: percentages.value.summaryChart
    },
    {
      name: $t('pricing.premium_product.timeframe_heatmap_chart_no_timeframe'),
      subtext: $t('pricing.premium_product.coming_soon'),
      available: true,
      percentage: percentages.value.heatmapChart
    }
  ]
})

const minorFeatures = computed<Feature[]>(() => {
  return [
    {
      name: $t('pricing.premium_product.no_ads'),
      available: props.product?.premium_perks.ad_free
    },
    {
      name: $t('pricing.premium_product.share_dashboard'),
      available: props.product?.premium_perks.share_custom_dashboards
    },
    {
      name: $t('pricing.premium_product.mobile_app_widget'),
      link: '/mobile',
      available: props.product?.premium_perks.mobile_app_widget
    },
    {
      name: $t('pricing.premium_product.manage_dashboard_via_api'),
      subtext: $t('pricing.premium_product.coming_soon'),
      available: props.product?.premium_perks.manage_dashboard_via_api
    }
  ]
})

</script>

<template>
  <div class="box-container" :popular="product.is_popular || null">
    <div class="name-container">
      <div class="name">
        {{ props.product?.product_name }}
      </div>
      <div v-if="product.is_popular" class="popular">
        {{ $t('pricing.premium_product.popular') }}
      </div>
    </div>
    <div class="features-container">
      <div class="prize">
        {{ prices.main }}
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ $t('pricing.per_month') }}</span><span v-if="!isYearly">*</span>
        </div>
        <div v-if="isYearly">
          {{ $t('pricing.amount_per_year', {amount: prices.yearly}) }}*
        </div>
      </div>
      <div v-if="isYearly" class="saving-info">
        <div>
          {{ $t('pricing.savings', {amount: prices.saving}) }}
        </div>
        <BcTooltip position="top" :fit-content="true">
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="saving-tooltip-container">
              {{ $t('pricing.savings_tooltip', {monthly: prices.monthly, monthly_yearly: prices.monthly_based_on_yearly}) }}
            </div>
          </template>
        </BcTooltip>
      </div>
      <div class="main-features-container">
        <PricingPremiumFeature
          v-for="feature in mainFeatures"
          :key="feature.name"
          :feature="feature"
        />
      </div>
      <div class="minor-features-container">
        <PricingPremiumFeature
          v-for="feature in minorFeatures"
          :key="feature.name"
          :feature="feature"
          :link="feature.link"
        />
      </div>
      <div v-if="planButton.isDowngrade" class="plan-button dismiss">
        {{ planButton.text }}
      </div>
      <Button v-else :label="planButton.text" class="plan-button" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  width: 400px;
  height: 100%;
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  background-color: var(--container-background);
  text-align: center;
  flex-shrink: 0;

  &[popular] {
    width: 460px;
    border-color: var(--primary-color);
  }

  .name-container {
    display: flex;
    justify-content: center;
    align-items: baseline;
    gap: 9px;
    padding: 18px 0;
    border-bottom: 2px solid var(--container-border-color);
    font-family: var(--montserrat-family);

    .name {
      font-size: 50px;
    }

    .popular {
      font-size: 35px;
      color: var(--primary-color);
    }
  }

  &[popular] .features-container {
    padding: 18px 65px;
  }

  &:not([popular]) .features-container {
    padding: 18px 35px;
  }

  .features-container {
    display: flex;
    flex-direction: column;
    font-family: var(--roboto-family);

    .prize {
      font-size: 70px;
      font-family: var(--montserrat-family);
    }

    .prize-subtext {
      color: var(--text-color-discreet);
      font-size: 21px;
      font-weight: 400;
      line-height: 1.85;
      display: flex;
      flex-direction: column;
      margin-bottom: 21px;
    }

    .saving-info {
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;
      gap: 13px;
      height: 37px;
      border-radius: 18px;
      background: var(--subcontainer-background);
      font-size: 17px;
      margin-bottom: 28px;
    }

    .main-features-container {
      display: flex;
      flex-direction: column;
      gap: 22px;
      margin-bottom: 35px;
    }

    .minor-features-container{
      display: flex;
      flex-direction: column;
      gap: 9px;
      margin-bottom: 35px;
    }
  }

  .plan-button {
    width: 100%;
    height: 53px;
    font-size: 25px;
    border-radius: 7px;
    margin-bottom: 5px;

    &.dismiss {
      display: flex;
      justify-content: center;
      align-items: center;
      cursor: pointer;
      color: var(--text-color-disabled);
    }
  }

  @media (max-width: 600px) {
    width: 240px;

    &[popular] {
      width: 320px;
    }

    .name-container {
      .name {
        font-size: 18px;
      }

      .popular {
        font-size: 16px;
      }
    }

    &[popular] .features-container {
      padding: 10px 60px;
    }

    &:not([popular]) .features-container {
      padding: 10px 20px;
    }

    .features-container {
      .prize {
        font-size: 30px;
      }

      .prize-subtext {
        font-size: 12px;
        margin-bottom: 18px;
        line-height: 1.4;
      }

      .saving-info {
        font-size: 10px;
        height: 21px;
        margin-bottom: 22px;
      }

      .main-features-container {
        gap: 15px;
        margin-bottom: 18px;
      }

      .minor-features-container {
        gap: 3px;
        margin-bottom: 18px;
      }
    }

    .plan-button {
      font-size: 14px;
      height: 30px;
    }
  }
}

.saving-tooltip-container {
  width: 150px;
  text-align: left;
}
</style>
