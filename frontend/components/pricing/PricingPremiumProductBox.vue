<script lang="ts" setup>

// TODO: Use format value for currency and normal numbers
// TODO: Tooltip icon is not visible right now and has been substituted with a simple "i"
// TODO: Mobile App Widget requires link
// TODO: Implement new feedback from PO (see txt)
// TODO: Orca box is higher than the others
// TODO: Fill bars based on Orca

import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumProduct } from '~/types/api/user'
import { formatTimeDuration } from '~/utils/format'

const { t } = useI18n()

interface Props {
  product?: PremiumProduct,
  isYearly?: boolean
}
const props = defineProps<Props>()

const monthlyPrice = computed(() => {
  if (props.isYearly) {
    return (props.product?.price_per_year_eur || 0) / 12 / 100
  }
  return (props.product?.price_per_month_eur || 0) / 100
})

const saving = computed(() => {
  return ((props.product?.price_per_month_eur || 0) * 12 - (props.product?.price_per_year_eur || 0)) / 100
})
</script>

<template>
  <div class="box-container">
    <div class="name-container">
      {{ props.product?.product_name }}
    </div>
    <div class="features-container">
      <div class="prize">
        €{{ monthlyPrice }}
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ t('pricing.premium_product.per_month') }}</span><span v-if="!isYearly">*</span>
        </div>
        <div v-if="isYearly">
          €{{ (product?.price_per_year_eur || 0) / 100 }} {{ t('pricing.premium_product.yearly') }}*
        </div>
      </div>
      <div v-if="isYearly" class="saving-info">
        <div>
          {{ t('pricing.premium_product.you_save', {amount: '€' + saving}) }}
        </div>
        <BcTooltip position="top" text="Compared to paying monthly, dawg.">
          <FontAwesomeIcon :icon="faInfoCircle" /> i
        </BcTooltip>
      </div>
      <div class="main-features-container">
        <BcPricingFeature
          :name="t('pricing.premium_product.validator_dashboards', {amount: product?.premium_perks.validator_dashboards}, (product?.premium_perks.validator_dashboards || 0) <= 1 ? 1 : 2)"
          :available="true"
          :bar-fill-percentage="50"
        />
        <BcPricingFeature
          :name="t('pricing.premium_product.validators_per_dashboard', {amount: product?.premium_perks.validators_per_dashboard})"
          :subtext="t('pricing.premium_product.per_validator', {amount: '€0.0899'})"
          :available="true"
          :bar-fill-percentage="50"
        />
        <BcPricingFeature
          :name="t('pricing.premium_product.timeframe_dashboard_chart', {timeframe: formatTimeDuration(product?.premium_perks.summary_chart_history_seconds, t)})"
          :available="true"
          :bar-fill-percentage="15"
        />
        <BcPricingFeature
          :name="t('pricing.premium_product.timeframe_heatmap_chart', {timeframe: formatTimeDuration(product?.premium_perks.heatmap_history_seconds, t)})"
          :available="true"
          :bar-fill-percentage="15"
        />
      </div>
      <div class="small-features-container">
        <BcPricingFeature :name="t('pricing.premium_product.no_ads')" :available="product?.premium_perks.ad_free" />
        <BcPricingFeature :name="t('pricing.premium_product.share_dashboard')" :available="product?.premium_perks.share_custom_dashboards" />
        <BcPricingFeature :name="t('pricing.premium_product.mobile_app_widget')" :available="product?.premium_perks.mobile_app_widget" />
        <BcPricingFeature
          :name="t('pricing.premium_product.manage_dashboard_via_api')"
          :subtext="product?.premium_perks.manage_dashboard_via_api ? '(' + t('pricing.premium_product.coming_soon') + ')' : undefined"
          :available="product?.premium_perks.manage_dashboard_via_api"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  width: 353px;
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
</style>
