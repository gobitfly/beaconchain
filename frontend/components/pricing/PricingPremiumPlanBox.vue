<script lang="ts" setup>

// TODO: Use format value for currency and normal numbers
// TODO: Tooltip icon is not visible right now and has been substituted with a simple "i"
// TODO: "Manage Dashboard via API" requries subtext for when it is available
// TODO: Mobile App Widget requires link
// TODO: Implement new feedback from PO (see txt)

import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumPlan } from '~/types/pricing'

const { t } = useI18n()

interface Props {
  plan?: PremiumPlan,
  isYearly?: boolean
}
const props = defineProps<Props>()
</script>

<template>
  <div class="box-container">
    <div class="name-container">
      {{ props.plan?.Name }}
    </div>
    <div class="features-container">
      <div class="prize">
        €9.99
      </div>
      <div class="prize-subtext">
        <div>
          <span>{{ t('pricing.plans.per_month') }}</span><span v-if="!isYearly">*</span>
        </div>
        <div v-if="isYearly">
          €1077,88 {{ t('pricing.plans.yearly') }}*
        </div>
      </div>
      <div v-if="isYearly" class="saving-info">
        <div>
          {{ t('pricing.plans.you_save', {amount: '€12'}) }}
        </div>
        <BcTooltip position="top" text="Compared to paying monthly, dawg.">
          <FontAwesomeIcon :icon="faInfoCircle" /> i
        </BcTooltip>
      </div>
      <div class="main-features-container">
        <BcPricingFeature
          :name="t('pricing.plans.validator_dashboards', {amount: plan?.ValidatorDashboards}, (plan?.ValidatorDashboards || 0) > 1 ? 0 : 1)"
          :available="true"
          :bar-fill-percentage="50"
        />
        <BcPricingFeature
          :name="t('pricing.plans.validators_per_dashboard', {amount: plan?.ValidatorsPerDashboard})"
          :subtext="t('pricing.plans.per_validator', {amount: '€0.0899'})"
          :available="true"
          :bar-fill-percentage="50"
        />
        <BcPricingFeature
          :name="t('pricing.plans.timeframe_dashboard_chart', {timeframe: '7 days'})"
          :available="true"
          :bar-fill-percentage="15"
        />
        <BcPricingFeature
          :name="t('pricing.plans.timeframe_heatmap_chart', {timeframe: '7 days'})"
          :available="true"
          :bar-fill-percentage="15"
        />
      </div>
      <div class="small-features-container">
        <BcPricingFeature :name="t('pricing.plans.no_ads')" :available="true" />
        <BcPricingFeature :name="t('pricing.plans.share_dashboard')" :available="true" />
        <BcPricingFeature :name="t('pricing.plans.mobile_app_widget')" :available="true" />
        <BcPricingFeature :name="t('pricing.plans.manage_dashboard_via_api')" :available="false" />
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
