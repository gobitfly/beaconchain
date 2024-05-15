<script lang="ts" setup>

// TODO: Use translations and mind singular/plural, relevant for example for "X Validator Dashboards"
// TODO: Tooltip icon is not visible right now and has been substituted with a simple "i"

import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type PremiumPlan } from '~/types/pricing'

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
        <div>per month</div>
        <div v-if="isYearly">
          €1077,88 yearly*
        </div>
      </div>
      <div v-if="isYearly" class="saving-info">
        <div>
          You save money!
        </div>
        <BcTooltip position="top" text="Compared to paying monthly, dawg.">
          <FontAwesomeIcon :icon="faInfoCircle" /> i
        </BcTooltip>
      </div>
      <BcPricingFeature name="1 Validator Dashboard" :available="true" :bar-fill-percentage="50" />
      <BcPricingFeature name="100 Validators per Dashboard" subtext="€0.0899 per validator" :available="true" :bar-fill-percentage="50" />
      <BcPricingFeature name="7 days dashboard chart history" :available="true" :bar-fill-percentage="15" />
      <BcPricingFeature name="7 days Heatmap chart history" :available="true" :bar-fill-percentage="15" />
      <BcPricingFeature name="No Ads" :available="true" />
      <BcPricingFeature name="Share Dashboard" :available="true" />
      <BcPricingFeature name="Mobile App Widget" :available="true" />
      <BcPricingFeature name="Manage Dashboard via API" :available="false" />
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
    gap: 15px;
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
    }
  }
}
</style>
