<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumProduct } from '~/types/api/user'
import { formatPremiumProductPrice } from '~/utils/format'
import type { Feature } from '~/types/pricing'

/// ///////////////
// import { formatTimeDuration } from '~/utils/format' TODO: See commented code below

const {
  bestPremiumProduct,
  currentPremiumSubscription,
  isPremiumSubscribedViaApp,
  products,
} = useProductsStore()
const { isLoggedIn } = useUserStore()
const { t: $t } = useTranslation()
const { isStripeDisabled, stripeCustomerPortal, stripePurchase } = useStripe()

interface Props {
  isYearly: boolean
  product: PremiumProduct
}
const props = defineProps<Props>()

const prices = computed(() => {
  const mainPrice = props.isYearly
    ? props.product.price_per_year_eur / 12
    : props.product.price_per_month_eur

  const savingAmount
    = props.product.price_per_month_eur * 12 - props.product.price_per_year_eur
  const savingDigits = savingAmount % 100 === 0 ? 0 : 2

  return {
    main: formatPremiumProductPrice($t, mainPrice),
    monthly: formatPremiumProductPrice($t, props.product.price_per_month_eur),
    monthly_based_on_yearly: formatPremiumProductPrice(
      $t,
      props.product.price_per_year_eur / 12,
    ),
    perValidator: formatPremiumProductPrice(
      $t,
      mainPrice
      / props.product.premium_perks.validators_per_dashboard
      / props.product.premium_perks.validator_dashboards,
      6,
    ),
    saving: formatPremiumProductPrice($t, savingAmount, savingDigits),
    yearly: formatPremiumProductPrice($t, props.product.price_per_year_eur),
  }
})

const percentages = computed(() => {
  if (bestPremiumProduct?.value === undefined) {
    return {
      heatmapChart: 100,
      summaryChart: 100,
      validatorDashboards: 100,
      validatorsPerDashboard: 100,
    }
  }

  const bestProduct = bestPremiumProduct.value
  let chartPercent = 1
  // TODO: remove check for chart_history_seconds once the API is live
  if (props.product.premium_perks.chart_history_seconds) {
    chartPercent
      = (props.product.premium_perks.chart_history_seconds.hourly
      / bestProduct.premium_perks.chart_history_seconds.hourly)
      * 100
  }
  return {
    heatmapChart: chartPercent,
    summaryChart: chartPercent,
    validatorDashboards:
      (props.product.premium_perks.validator_dashboards
      / bestProduct.premium_perks.validator_dashboards)
      * 100,
    validatorsPerDashboard:
      (props.product.premium_perks.validators_per_dashboard
      / bestProduct.premium_perks.validators_per_dashboard)
      * 100,
  }
})

async function buttonCallback() {
  if (planButton.value.disabled) {
    return
  }

  if (isLoggedIn.value) {
    if (currentPremiumSubscription.value) {
      await stripeCustomerPortal()
    }
    else {
      await stripePurchase(
        props.isYearly
          ? props.product.stripe_price_id_yearly
          : props.product.stripe_price_id_monthly,
        1,
      )
    }
  }
  else {
    await navigateTo('/login')
  }
}

const planButton = computed(() => {
  let isDowngrade = false
  let text = $t('pricing.premium_product.button.select_plan')

  if (isLoggedIn.value) {
    if (currentPremiumSubscription.value) {
      const subscribedProduct = products.value?.premium_products.find(
        product =>
          product.product_id_monthly
          === currentPremiumSubscription.value!.product_id
          || product.product_id_yearly
          === currentPremiumSubscription.value!.product_id,
      )
      if (
        currentPremiumSubscription.value.product_id
        === props.product.product_id_monthly
        || currentPremiumSubscription.value.product_id
        === props.product.product_id_yearly
        || subscribedProduct === undefined
      ) {
        // (this box is either for the subscribed product)
        // || (the user has an unknown product, possible from V1 or maybe a custom plan)
        text = $t('pricing.premium_product.button.manage_plan')
      }
      else if (
        subscribedProduct.price_per_month_eur
        < props.product.price_per_month_eur
      ) {
        text = $t('pricing.premium_product.button.upgrade')
      }
      else {
        isDowngrade = true
        text = $t('pricing.premium_product.button.downgrade')
      }
    }
  }
  else {
    text = $t('pricing.get_started')
  }

  const disabled
    = isStripeDisabled.value || isPremiumSubscribedViaApp.value || undefined

  return { disabled, isDowngrade, text }
})

const mainFeatures = computed<Feature[]>(() => {
  const validatorPerDashboardAmount = formatNumber(
    props.product?.premium_perks.validators_per_dashboard,
  )
  return [
    {
      available: true,
      name: $t(
        'pricing.premium_product.validator_dashboards',
        {
          amount: formatNumber(
            props.product?.premium_perks.validator_dashboards,
          ),
        },
        (props.product?.premium_perks.validator_dashboards || 0) <= 1 ? 1 : 2,
      ),
      percentage: percentages.value.validatorDashboards,
    },
    {
      available: true,
      name:
        props.product.premium_perks.validator_dashboards === 1
          ? $t('pricing.premium_product.validators', {
            amount: validatorPerDashboardAmount,
          })
          : $t('pricing.premium_product.validators_per_dashboard', {
            amount: validatorPerDashboardAmount,
          }),
      percentage: percentages.value.validatorsPerDashboard,
      subtext: $t('pricing.per_validator', {
        amount: prices.value.perValidator,
      }),
      tooltip: $t('pricing.pectra_tooltip', {
        effectiveBalance: formatNumber(
          props.product?.premium_perks.validators_per_dashboard * 32,
        ),
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
      available: props.product?.premium_perks.ad_free,
      name: $t('pricing.premium_product.no_ads'),
    },
    {
      available: props.product?.premium_perks.share_custom_dashboards,
      name: $t('pricing.premium_product.share_dashboard'),
    },
    {
      available: props.product?.premium_perks.mobile_app_widget,
      link: '/mobile',
      name: $t('pricing.premium_product.mobile_app_widget'),
    },
    {
      available: props.product?.premium_perks.manage_dashboard_via_api,
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
      <div class="name">
        {{ props.product?.product_name }}
      </div>
      <div
        v-if="product.is_popular"
        class="popular"
      >
        {{ $t("pricing.premium_product.popular") }}
      </div>
    </div>
    <div class="features-container">
      <div class="prize">
        {{ prices.main }}
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ $t("pricing.per_month") }}</span><span v-if="!isYearly">*</span>
        </div>
        <div v-if="isYearly">
          {{ $t("pricing.amount_per_year", { amount: prices.yearly }) }}*
        </div>
      </div>
      <div
        v-if="isYearly"
        class="saving-info"
      >
        <div>
          {{ $t("pricing.savings", { amount: prices.saving }) }}
        </div>
        <BcTooltip
          position="top"
          :fit-content="true"
        >
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="saving-tooltip-container">
              {{
                $t("pricing.savings_tooltip", {
                  monthly: prices.monthly,
                  monthly_yearly: prices.monthly_based_on_yearly,
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
      <div
        v-if="planButton.isDowngrade"
        :disabled="planButton.disabled"
        class="plan-button dismiss"
        @click="buttonCallback()"
      >
        {{ planButton.text }}
      </div>
      <Button
        v-else
        :label="planButton.text"
        :disabled="planButton.disabled"
        class="plan-button"
        @click="buttonCallback()"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/pricing.scss";

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

  .plan-button {
    width: 100%;
    @include pricing.pricing_button;

    &.dismiss {
      display: flex;
      justify-content: center;
      align-items: center;
      cursor: pointer;
      color: var(--text-color-disabled);
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
  }
}

.saving-tooltip-container {
  width: 150px;
  text-align: left;
}
</style>
