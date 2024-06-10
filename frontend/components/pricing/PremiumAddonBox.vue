<script lang="ts" setup>

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { type ExtraDashboardValidatorsPremiumAddon, ProductCategoryPremiumAddon } from '~/types/api/user'
import { formatPremiumProductPrice } from '~/utils/format'

const { t: $t } = useI18n()
const { user, isLoggedIn } = useUserStore()
const { stripeCustomerPortal, stripePurchase, isStripeDisabled } = useStripe()

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
    perValidator: $t('pricing.per_validator', { amount: prices.value.perValidator })
  }
})

const addonSubscription = computed(() => {
  return user.value?.subscriptions?.find(sub => sub.product_category === ProductCategoryPremiumAddon)
})

// TODO: Ponder on moving this to provider (as the code for the plans is very similar)
async function buttonCallback () {
  if (isStripeDisabled.value) {
    return
  }

  if (isLoggedIn.value) {
    if (addonSubscription.value) {
      await stripeCustomerPortal()
    } else {
      await stripePurchase(props.isYearly ? props.addon.price_per_year_eur : props.addon.price_per_month_eur, 1)
    }
  } else {
    await navigateTo('/register')
  }
}

const addonButton = computed(() => {
  return {
    text: addonSubscription.value ? $t('pricing.addons.button.manage_addon') : $t('pricing.addons.button.select_addon'),
    disabled: isStripeDisabled.value
  }
})

</script>

<template>
  <div class="box-container">
    <div class="summary-container">
      <div class="validator-count">
        {{ text.validatorCount }}
        <div class="subtext">
          {{ $t('pricing.addons.per_dashboard') }}
          <BcTooltip position="top" :fit-content="true">
            <FontAwesomeIcon :icon="faInfoCircle" class="tooltip-icon" />
            <template #tooltip>
              <div class="saving-tooltip-container">
                {{ $t('pricing.pectra_tooltip', { effectiveBalance: formatNumber(props.addon?.extra_dashboard_validators * 32) }) }}
              </div>
            </template>
          </BcTooltip>
        </div>
        <div class="per-validator">
          {{ text.perValidator }}
        </div>
      </div>
    </div>
    <div class="description-container">
      <div class="price">
        <template v-if="isYearly">
          <div>
            {{ prices.monthly_based_on_yearly }}
          </div>
          <div class="month" yearly>
            {{ $t('pricing.per_month') }}
          </div>
          <div class="year">
            {{ $t('pricing.amount_per_year', {amount: prices.yearly}) }}*
          </div>
        </template>
        <template v-else>
          <div>
            {{ prices.monthly }}
          </div>
          <div class="month">
            {{ $t('pricing.per_month') }}*
          </div>
        </template>
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
      <Button :label="addonButton.text" :disabled="addonButton.disabled" class="select-button" @click="buttonCallback" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/pricing.scss';

.box-container {
  width: 290px;
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
    padding: 28px 0 21px 0;

    .validator-count {
      font-size: 20px;
      font-weight: 600;

      .subtext {
        font-weight: 400;
        margin-bottom: 16px;

        .tooltip-icon {
          width: 15px;
        }
      }

      .per-validator {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 400;
      }
    }
  }

  .description-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 28px 29px 28px;

    .price {
      font-size: 26px;
      font-weight: 600;
      margin-bottom: 24px;

      .month {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 600;

        &[yearly] {
          font-size: 14px;
          font-weight: 500;
        }
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
      height: 30px;
      border-radius: 15px;
      background: var(--subcontainer-background);
      font-size: 15px;
      margin-bottom: 56px;
    }

    .select-button {
      width: 100%;
      @include pricing.pricing_button;
    }
  }

  @media (max-width: 1360px) {
    width: 200px;

    .summary-container {
      padding: 20px 0 18px 0;

      .validator-count {
        font-size: 14px;

        .subtext {
          margin-bottom: 10px;

          .tooltip-icon {
            width: 13px;
          }
        }

        .per-validator {
          font-size: 12px;
        }
      }
    }

    .description-container {
      padding: 10px 25px 10px 25px;

      .price {
        font-size: 18px;
        margin-bottom: 17px;

        .month {
          font-size: 12px;

          &[yearly] {
            font-size: 10px;
          }
        }

        .year {
          font-size: 10px;
        }
      }

      .saving-info {
        height: 21px;
        gap: 4px;
        font-size: 10px;
        margin-bottom: 39px;
      }

      .select-button {
        padding-left: 10px;
        padding-right: 10px;
      }
    }
  }
}

.saving-tooltip-container {
  width: 150px;
  text-align: left;
}
</style>
