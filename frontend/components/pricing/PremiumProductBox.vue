<script lang="ts" setup>
// TODO: Add Select Plan button

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumProduct } from '~/types/api/user'
import { formatFiat } from '~/utils/format'
// import { formatTimeDuration } from '~/utils/format' TODO: See commented code below

const { t } = useI18n()

interface Props {
  product: PremiumProduct,
  compareProduct?: PremiumProduct,
  isYearly: boolean
}
const props = defineProps<Props>()

const formatPremiumProductPrice = (price: number, digits?: number) => {
  return formatFiat(price / 100, 'EUR', t('locales.currency'), digits ?? 2, digits ?? 2)
}

const prices = computed(() => {
  const mainPrice = props.isYearly ? props.product.price_per_year_eur / 12 : props.product.price_per_month_eur

  const savingAmount = props.product.price_per_month_eur * 12 - props.product.price_per_year_eur
  const savingDigits = savingAmount % 100 === 0 ? 0 : 2

  return {
    main: formatPremiumProductPrice(mainPrice),
    monthly: formatPremiumProductPrice(props.product.price_per_month_eur),
    monthly_based_on_yearly: formatPremiumProductPrice(props.product.price_per_year_eur / 12),
    yearly: formatPremiumProductPrice(props.product.price_per_year_eur),
    saving: formatPremiumProductPrice(savingAmount, savingDigits),
    perValidator: formatPremiumProductPrice(mainPrice / props.product.premium_perks.validators_per_dashboard, 5)
  }
})

const barFillPercentages = computed(() => {
  return {
    validatorDashboards: props.product.premium_perks.validator_dashboards / (props.compareProduct?.premium_perks.validator_dashboards ?? 1) * 100,
    validatorsPerDashboard: props.product.premium_perks.validators_per_dashboard / (props.compareProduct?.premium_perks.validators_per_dashboard ?? 1) * 100,
    summaryChart: props.product.premium_perks.summary_chart_history_seconds / (props.compareProduct?.premium_perks.summary_chart_history_seconds ?? 1) * 100,
    heatmapChart: props.product.premium_perks.heatmap_history_seconds / (props.compareProduct?.premium_perks.heatmap_history_seconds ?? 1) * 100
  }
})
</script>

<template>
  <div class="box-container">
    <div class="name-container">
      {{ props.product?.product_name }}
    </div>
    <div class="features-container">
      <div class="prize">
        {{ prices.main }}
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ t('pricing.premium_product.per_month') }}</span><span v-if="!isYearly">*</span>
        </div>
        <div v-if="isYearly">
          {{ prices.yearly }} {{ t('pricing.premium_product.yearly') }}*
        </div>
      </div>
      <div v-if="isYearly" class="saving-info">
        <div>
          {{ t('pricing.premium_product.savings', {amount: prices.saving}) }}
        </div>
        <BcTooltip position="top" :fit-content="true">
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="saving-tooltip-container">
              {{ t('pricing.premium_product.savings_tooltip', {monthly: prices.monthly, monthly_yearly: prices.monthly_based_on_yearly}) }}
            </div>
          </template>
        </BcTooltip>
      </div>
      <div class="main-features-container">
        <PricingPremiumFeature
          :name="t('pricing.premium_product.validator_dashboards', {amount: product?.premium_perks.validator_dashboards}, (product?.premium_perks.validator_dashboards || 0) <= 1 ? 1 : 2)"
          :available="true"
          :bar-fill-percentage="barFillPercentages.validatorDashboards"
        />
        <PricingPremiumFeature
          :name="t('pricing.premium_product.validators_per_dashboard', {amount: product?.premium_perks.validators_per_dashboard})"
          :subtext="t('pricing.premium_product.per_validator', {amount: prices.perValidator})"
          :available="true"
          :bar-fill-percentage="barFillPercentages.validatorsPerDashboard"
        />
        <!--
          TODO: For now we hide the number until the backend knows what it is capable of
          :name="t('pricing.premium_product.timeframe_dashboard_chart', {timeframe: formatTimeDuration(product?.premium_perks.summary_chart_history_seconds, t)})"
        -->
        <PricingPremiumFeature
          :name="t('pricing.premium_product.timeframe_dashboard_chart_no_timeframe')"
          :subtext="t('pricing.premium_product.coming_soon')"
          :available="true"
          :bar-fill-percentage="barFillPercentages.summaryChart"
        />
        <!--
          TODO: For now we hide the number until the backend knows what it is capable of
          :name="t('pricing.premium_product.timeframe_heatmap_chart', {timeframe: formatTimeDuration(product?.premium_perks.heatmap_history_seconds, t)})"
        -->
        <PricingPremiumFeature
          :name="t('pricing.premium_product.timeframe_heatmap_chart_no_timeframe')"
          :subtext="t('pricing.premium_product.coming_soon')"
          :available="true"
          :bar-fill-percentage="barFillPercentages.heatmapChart"
        />
      </div>
      <div class="small-features-container">
        <PricingPremiumFeature :name="t('pricing.premium_product.no_ads')" :available="product?.premium_perks.ad_free" />
        <PricingPremiumFeature :name="t('pricing.premium_product.share_dashboard')" :available="product?.premium_perks.share_custom_dashboards" />
        <PricingPremiumFeature
          :name="t('pricing.premium_product.mobile_app_widget')"
          link="/mobile"
          :available="product?.premium_perks.mobile_app_widget"
        />
        <PricingPremiumFeature
          :name="t('pricing.premium_product.manage_dashboard_via_api')"
          :subtext="t('pricing.premium_product.coming_soon')"
          :available="product?.premium_perks.manage_dashboard_via_api"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  width: 353px;
  height: 100%;
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  text-align: center;

  .name-container {
    font-size: 50px;
    padding: 18px 0;
    border-bottom: 2px solid var(--container-border-color);
  }

  .features-container {
    display: flex;
    flex-direction: column;
    padding: 18px 35px;

    .prize {
      font-size: 70px;
    }

    .prize-subtext {
      color: var(--text-color-discreet);
      font-size: 22px;
      font-weight: 400;
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

    .small-features-container{
      display: flex;
      flex-direction: column;
      gap: 9px;
    }
  }
}

.saving-tooltip-container {
  width: 150px;
  text-align: left;
}
</style>
