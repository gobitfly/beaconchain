<script lang="ts" setup>

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle, faMinus, faPlus } from '@fortawesome/pro-regular-svg-icons'
import { type ExtraDashboardValidatorsPremiumAddon, ProductCategoryPremiumAddon } from '~/types/api/user'
import { formatPremiumProductPrice } from '~/utils/format'
import { Target } from '~/types/links'

const { t: $t } = useI18n()
const { user, isLoggedIn } = useUserStore()
const { stripeCustomerPortal, stripePurchase, isStripeDisabled } = useStripe()

interface Props {
  addon: ExtraDashboardValidatorsPremiumAddon,
  isYearly: boolean,
  maximumValidatorLimit?: number
}
const props = defineProps<Props>()

const quantityForPurchase = ref(1)

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

const boxText = computed(() => {
  return {
    validatorCount: $t('pricing.addons.validator_amount', { amount: formatNumber(props.addon.extra_dashboard_validators) }),
    perValidator: $t('pricing.per_validator', { amount: prices.value.perValidator })
  }
})

const addonSubscriptionCount = computed(() => {
  return user.value?.subscriptions?.filter(sub => sub.product_category === ProductCategoryPremiumAddon && (sub.product_id === props.addon.product_id_monthly || sub.product_id === props.addon.product_id_yearly)).length || 0
})

const addonButton = computed(() => {
  let text = $t('pricing.get_started')
  if (isLoggedIn.value) {
    text = addonSubscriptionCount.value > 0 ? $t('pricing.addons.button.manage_addon') : $t('pricing.addons.button.select_addon')
  }

  async function callback () {
    if (isStripeDisabled.value) {
      return
    }

    if (isLoggedIn.value) {
      if (addonSubscriptionCount.value > 0) {
        await stripeCustomerPortal()
      } else {
        await stripePurchase(props.isYearly ? props.addon.stripe_price_id_yearly : props.addon.stripe_price_id_monthly, quantityForPurchase.value)
      }
    } else {
      await navigateTo('/login')
    }
  }

  return {
    text,
    disabled: isStripeDisabled.value,
    callback
  }
})

const maximumQuantity = computed(() => {
  return Math.floor(((props.maximumValidatorLimit || 10000) - (user.value?.premium_perks.validators_per_dashboard || 0)) / props.addon.extra_dashboard_validators)
})

const limitReached = computed(() => {
  return quantityForPurchase.value >= maximumQuantity.value
})

const purchaseQuantityButtons = computed(() => {
  return {
    minus: {
      disabled: quantityForPurchase.value <= 1,
      callback: () => {
        if (quantityForPurchase.value > 1) {
          quantityForPurchase.value--
        }
      }
    },
    plus: {
      disabled: limitReached.value,
      callback: () => {
        if (quantityForPurchase.value < maximumQuantity.value) {
          quantityForPurchase.value++
        }
      }
    }
  }
})

</script>

<template>
  <div class="box-container">
    <div class="summary-container">
      <div class="validator-count">
        {{ boxText.validatorCount }}
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
          {{ boxText.perValidator }}
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
      <div class="quantity-row">
        <div v-if="addonSubscriptionCount" class="quantity-label">
          {{ $t('pricing.addons.currently_active', { amount: addonSubscriptionCount }) }}
        </div>
        <div v-else class="quantity-setter">
          <Button
            class="p-button-icon-only"
            :disabled="purchaseQuantityButtons.minus.disabled"
            @click="purchaseQuantityButtons.minus.callback"
          >
            <FontAwesomeIcon :icon="faMinus" />
          </Button>
          <InputNumber
            v-model="quantityForPurchase"
            class="quantity-input"
            input-id="integeronly"
            :min="1"
            :max="maximumQuantity"
          />
          <Button
            class="p-button-icon-only"
            :disabled="purchaseQuantityButtons.plus.disabled"
            @click="purchaseQuantityButtons.plus.callback"
          >
            <FontAwesomeIcon :icon="faPlus" />
          </Button>
        </div>
      </div>
      <div class="limit-reached-row">
        <div v-if="limitReached">
          {{ tOf($t, 'pricing.addons.contact_support', 0) }}
          <BcLink to="https://dsc.gg/beaconchain  " :target="Target.External" class="link">
            {{ tOf($t, 'pricing.addons.contact_support', 1) }}
          </BcLink>
          {{ tOf($t, 'pricing.addons.contact_support', 2) }}
        </div>
      </div>
      <Button :label="addonButton.text" :disabled="addonButton.disabled" class="select-button" @click="addonButton.callback" />
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
      margin-bottom: 24px;
    }

    .quantity-row {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 30px;

      .quantity-label {
        font-size: 17px;
      }

      .quantity-setter {
        height: 100%;
        display: flex;
        justify-content: center;
        gap: 15px;

        .quantity-input {
          width: 45px;

          > :first-child {
            width: 100%;
            text-align: center;
          }
        }

        > * {
          height: 100%;
        }
      }

      margin-bottom: 20px;
    }

    .limit-reached-row {
      height: 16px;
      font-size: 13px;

      margin-bottom: 20px;
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
        margin-bottom: 17px;
      }

      .quantity-row {
        height: 20px;

        .quantity-label {
          font-size: 12px;
        }

        .quantity-setter {
          gap: 8px;

          .quantity-input {
            width: 35px;

            > :first-child {
              font-size: 12px;
            }
          }

          > .p-button {
            width: 20px;
          }
        }

        margin-bottom: 10px;
      }

      .limit-reached-row {
        height: 10px;
        font-size: 8px;

        margin-bottom: 20px;
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
