<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { PremiumProduct } from '~/types/api/user'
import type { Feature } from '~/types/pricing'

const {
  isPaymentYearly,
  product,
} = defineProps<{
  isPaymentYearly: boolean,
  product: PremiumProduct,
}>()

const {
  bestPremiumProduct,
  currentPremiumSubscription,
  isPremiumSubscribedViaApp,
  products,
} = useProductsStore()
const { isLoggedIn } = useUserStore()
const { t: $t } = useTranslation()
const { promoCode } = usePromoCode()
const {
  isStripeDisabled,
  stripeCustomerPortal,
  stripePurchase,
} = useStripe()
const {
  displayCurrencyDefault,
  formatAmount,
} = useCurrency()

const productPrice = computed(() => {
  return isPaymentYearly
    ? product.price_per_year_eur / 12
    : product.price_per_month_eur
})

const getOldMaxEffectiveBalance = (
  hasCurrencyDisplay: boolean = false,
  targetUnit: CryptoUnit = 'wei',
) => formatAmount('32', {
  hasCurrencyDisplay,
  maximumFractionDigits: 0,
  minimumFractionDigits: 0,
  sourceUnit: 'base',
  targetCurrency: displayCurrencyDefault.main,
  targetUnit,
  useGrouping: false,
})

const oldMaxEffectiveBalance = Number(getOldMaxEffectiveBalance())
const oldMaxEffectiveBalanceWithUnit = getOldMaxEffectiveBalance(true, 'base')

// we don't charge per Validator anymore, but this is used to show users that the price
// they used to pay per Validator hasn't changed now that we charge by Effective Balance
const pricePerValidator = computed(() => {
  return (productPrice.value * oldMaxEffectiveBalance)
    / product.premium_perks.effective_balance_per_dashboard
    / product.premium_perks.validator_dashboards
})

const totalYearlyPricePerMonth = computed(() => {
  return product.price_per_year_eur / 12
})

const yearlySubscriptionSavings = computed(() => {
  return (product.price_per_month_eur * 12 - product.price_per_year_eur)
})

const percentages = computed(() => {
  if (bestPremiumProduct?.value === undefined) {
    return {
      effectiveBalancePerDashboard: 100,
      heatmapChart: 100,
      summaryChart: 100,
      validatorDashboards: 100,
    }
  }

  const bestProduct = bestPremiumProduct.value
  let chartPercent = 1
  // TODO: remove check for chart_history_seconds once the API is live
  if (product.premium_perks.chart_history_seconds) {
    chartPercent
      = (product.premium_perks.chart_history_seconds.hourly
        / bestProduct.premium_perks.chart_history_seconds.hourly)
      * 100
  }
  return {
    effectiveBalancePerDashboard:
      (product.premium_perks.effective_balance_per_dashboard
        / bestProduct.premium_perks.effective_balance_per_dashboard)
      * 100,
    heatmapChart: chartPercent,
    summaryChart: chartPercent,
    validatorDashboards:
      (product.premium_perks.validator_dashboards
        / bestProduct.premium_perks.validator_dashboards)
      * 100,
  }
})

async function handleProductPurchase() {
  if (isLoggedIn.value) {
    if (currentPremiumSubscription.value) {
      await stripeCustomerPortal()
    }
    else {
      await stripePurchase(
        isPaymentYearly
          ? product.stripe_price_id_yearly
          : product.stripe_price_id_monthly,
        1,
      )
    }
  }
  else {
    await navigateTo({
      path: '/login', query: { promoCode },
    })
  }
}

const planButton = computed(() => {
  let isDowngrade = false
  let text = $t('pricing.premium_product.button.select_plan')

  if (currentPremiumSubscription.value) {
    const subscribedProduct = products.value?.premium_products.find(
      product =>
        product.product_id_monthly
        === currentPremiumSubscription.value!.product_id || product.product_id_yearly
        === currentPremiumSubscription.value!.product_id,
    )

    if (
      currentPremiumSubscription.value.product_id
      === product.product_id_monthly || currentPremiumSubscription.value.product_id
      === product.product_id_yearly || subscribedProduct === undefined
    ) {
      // (this box is either for the subscribed product)
      // || (the user has an unknown product, possible from V1 or maybe a custom plan)
      text = $t('pricing.premium_product.button.manage_plan')
    }
    else if (
      subscribedProduct.price_per_month_eur
      < product.price_per_month_eur
    ) {
      text = $t('pricing.premium_product.button.upgrade')
    }
    else {
      isDowngrade = true
      text = $t('pricing.premium_product.button.downgrade')
    }
  }
  else {
    text = $t('pricing.get_started')
  }

  return {
    isDowngrade,
    text,
  }
})

const isSubmitButtonDisabled = computed(() => {
  return isStripeDisabled.value || isPremiumSubscribedViaApp.value || undefined
})

const mainFeatures = computed<Feature[]>(() => {
  const maxDashboardEffectiveBalance = formatAmount(`${product.premium_perks.effective_balance_per_dashboard}`, {
    minimumFractionDigits: 0,
    sourceUnit: 'wei',
    targetCurrency: displayCurrencyDefault.main,
  })

  return [
    {
      available: true,
      name: $t(
        'pricing.premium_product.validator_dashboards',
        {
          amount: formatNumber(
            product?.premium_perks.validator_dashboards,
          ),
        },
        (product?.premium_perks.validator_dashboards || 0) <= 1 ? 1 : 2,
      ),
      percentage: percentages.value.validatorDashboards,
    },
    {
      available: true,
      name: $t('pricing.premium_product.max_effective_balance', { amount: maxDashboardEffectiveBalance }),
      percentage: percentages.value.effectiveBalancePerDashboard,
      subtext: $t('pricing.per_min_validator_deposit', {
        amount: formatFiatCurrency(pricePerValidator.value, {
          minimumFractionDigits: 6,
        }),
        old_validator_max_effective_balance: oldMaxEffectiveBalanceWithUnit,
      }),
      tooltip: $t('pricing.premium_product.max_effective_balance_tooltip', {
        amount: maxDashboardEffectiveBalance,
      }),
    },
    {
      available: true,
      name: $t(
        'pricing.premium_product.timeframe_dashboard_chart_no_timeframe',
      ),
      percentage: percentages.value.summaryChart,
      subtext: $t('pricing.premium_product.coming_soon'),
    },
    {
      available: true,
      name: $t('pricing.premium_product.timeframe_heatmap_chart_no_timeframe'),
      percentage: percentages.value.heatmapChart,
      subtext: $t('pricing.premium_product.coming_soon'),
    },
  ]
})

const minorFeatures = computed<Feature[]>(() => {
  return [
    {
      available: product?.premium_perks.ad_free,
      name: $t('pricing.premium_product.no_ads'),
    },
    {
      available: product?.premium_perks.share_custom_dashboards,
      name: $t('pricing.premium_product.share_dashboard'),
    },
    {
      available: product?.premium_perks.mobile_app_widget,
      link: '/mobile',
      name: $t('pricing.premium_product.mobile_app_widget'),
    },
    {
      available: product?.premium_perks.manage_dashboard_via_api,
      name: $t('pricing.premium_product.manage_dashboard_via_api'),
      subtext: $t('pricing.premium_product.coming_soon'),
    },
  ]
})
</script>

<template>
  <div
    class="box-container"
    :popular="product.is_popular || null"
  >
    <div class="name-container">
      <h3 class="name">
        {{ product?.product_name }}
      </h3>
      <div
        v-if="product.is_popular"
        class="popular"
      >
        {{ $t("pricing.premium_product.popular") }}
      </div>
    </div>
    <div class="features-container">
      <div class="prize">
        {{ formatFiatCurrency(productPrice) }}
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ $t("pricing.per_month") }}</span>
          <span v-if="!isPaymentYearly">*</span>
        </div>
        <div v-if="isPaymentYearly">
          {{ $t("pricing.amount_per_year", { amount: formatFiatCurrency(product.price_per_year_eur) }) }}*
        </div>
      </div>
      <div
        v-if="isPaymentYearly"
        class="info-badge"
      >
        {{ $t("pricing.savings", { amount: formatFiatCurrency(yearlySubscriptionSavings) }) }}
        <BcTooltip
          position="top"
          :fit-content="true"
        >
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="saving-tooltip-container">
              {{
                $t("pricing.savings_tooltip", {
                  monthly: formatFiatCurrency(product.price_per_month_eur),
                  monthly_yearly: formatFiatCurrency(totalYearlyPricePerMonth),
                })
              }}
            </div>
          </template>
        </BcTooltip>
      </div>
      <div class="main-features-container">
        <PricingPremiumFeature
          v-for="feature in mainFeatures"
          :key="feature.name"
          :feature
        />
      </div>
      <div class="minor-features-container">
        <PricingPremiumFeature
          v-for="feature in minorFeatures"
          :key="feature.name"
          :feature
          :link="feature.link"
        />
      </div>
      <BcButton
        :label="planButton.text"
        :is-disabled="isSubmitButtonDisabled"
        :class="{ 'submit-button--downgrade': planButton.isDowngrade }"
        @click="handleProductPurchase()"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  box-sizing: border-box;
  width: 293px;
  height: 100%;
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  background-color: var(--container-background);
  text-align: center;
  flex-shrink: 0;

  &[popular] {
    width: 381px;
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
      font-size: 41px;
    }

    .popular {
      font-size: 29px;
      color: var(--primary-color);
    }
  }

  &[popular] .features-container {
    padding: 18px 64px 29px 64px;
  }

  &:not([popular]) .features-container {
    padding: 18px 25px 29px 25px;
  }

  .features-container {
    display: flex;
    flex-direction: column;
    font-family: var(--roboto-family);

    .prize {
      font-size: 59px;
      font-family: var(--montserrat-family);
    }

    .prize-subtext {
      color: var(--text-color-discreet);
      font-size: 18px;
      font-weight: 400;
      line-height: 1.85;
      display: flex;
      flex-direction: column;
      margin-bottom: 18px;
    }

    .saving-info {
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;
      gap: 13px;
      height: 30px;
      border-radius: 18px;
      background: var(--subcontainer-background);
      font-size: 15px;
      margin-bottom: 28px;
    }

    .main-features-container {
      display: flex;
      flex-direction: column;
      gap: 22px;
      margin-bottom: 35px;
    }

    .minor-features-container {
      display: flex;
      flex-direction: column;
      gap: 9px;
      margin-bottom: 35px;
    }
  }

  @media (max-width: 1360px) {
    width: 240px;

    &[popular] {
      width: 300px;
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
      padding: 10px 45px;
    }

    &:not([popular]) .features-container {
      padding: 10px 18px;
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

      .saving-tooltip-container {
        width: 150px;
        text-align: left;
      }

      .info-badge {
        background-color: var(--subcontainer-background);
        font-size: 0.7rem;
        padding: 0.25rem 0.15rem;
        border-radius: 4px;
        gap: 0.25rem;
        justify-content: center;
        margin-bottom: 0.75rem;
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
  }

  .submit-button--downgrade {
    background-color: transparent;
    color: var(--text-color-discreet);
    border: none;

    &:hover {
      background-color: transparent !important;
      color: var(--text-color-discreet) !important;
      border: none !important;
    }
  }
}
</style>
