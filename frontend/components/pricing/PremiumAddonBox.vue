<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faInfoCircle,
  faMinus,
  faPlus,
} from '@fortawesome/pro-regular-svg-icons'
import {
  type ExtraDashboardValidatorsPremiumAddon,
  ProductCategoryPremiumAddon,
} from '~/types/api/user'

const {
  addon,
  effectiveBalancePerDashboardLimit,
  isPaymentYearly,
} = defineProps<({
  addon: ExtraDashboardValidatorsPremiumAddon,
  effectiveBalancePerDashboardLimit: string,
  isPaymentYearly: boolean,
})>()

const { t: $t } = useTranslation()
const {
  displayCurrencyDefault,
  formatAmount,
} = useCurrency()

const {
  isLoggedIn,
  premium_perks,
  user,
} = useUserStore()

const quantity = ref(1)

const pricePerUnit = computed(() => {
  return isPaymentYearly
    ? addon.price_per_year_eur / 12
    : addon.price_per_month_eur
})

const totalMonthlyPrice = computed(() => {
  return addon.price_per_month_eur * quantity.value
})

const totalYearlyPricePerMonth = computed(() => {
  return (addon.price_per_year_eur / 12) * quantity.value
})

const totalYearlyPrice = computed(() => {
  return addon.price_per_year_eur * quantity.value
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

// This is used to show users that the price they used to pay per Validator
// hasn't changed now that we charge by Effective Balance
const pricePerValidator = computed(() => {
  return divideBigNumbers(
    (pricePerUnit.value * oldMaxEffectiveBalance),
    addon.extra_dashboard_effective_balance,
  )
})

const yearlySubscriptionSavings = computed(() => {
  return (addon.price_per_month_eur * 12 - addon.price_per_year_eur)
    * quantity.value
})

const extraEffectiveBalance = computed(() =>
  formatAmount(`${addon.extra_dashboard_effective_balance}`, {
    minimumFractionDigits: 0,
    targetCurrency: displayCurrencyDefault.main,

  }),
)

const addonSubscriptionCount = computed(() => {
  return (
    user.value?.subscriptions?.filter(
      sub =>
        sub.product_category === ProductCategoryPremiumAddon
        && (sub.product_id === addon.product_id_monthly
          || sub.product_id === addon.product_id_yearly),
    ).length || 0
  )
})

const isQuantityLimitReached = computed(() => {
  return quantity.value >= maximumQuantity.value
})

const maximumQuantity = computed(() => {
  const unusedDashboardEffectiveBalance = addBigNumbers(
    effectiveBalancePerDashboardLimit ?? 0,
    -(premium_perks.value?.effective_balance_per_dashboard ?? 0),
  )
  const extraAddonEffectiveBalanacePerDashboard = addon.extra_dashboard_effective_balance

  return Math.floor(
    Number(divideBigNumbers(unusedDashboardEffectiveBalance, extraAddonEffectiveBalanacePerDashboard)),
  )
})

const {
  isStripeDisabled,
  stripeCustomerPortal,
  stripePurchase,
} = useStripe()
const { promoCode } = usePromoCode()

const isDisabledSubmitButton = computed(() =>
  isStripeDisabled.value
  || quantity.value > maximumQuantity.value
  || quantity.value < 1,
)

const handleSubmitPurchase = async () => {
  if (isStripeDisabled.value) {
    return
  }

  if (isLoggedIn.value) {
    if (addonSubscriptionCount.value > 0) {
      await stripeCustomerPortal()
    }
    else {
      await stripePurchase(
        isPaymentYearly
          ? addon.stripe_price_id_yearly
          : addon.stripe_price_id_monthly,
        quantity.value,
      )
    }
  }
  else {
    await navigateTo({
      path: '/login', query: { promoCode },
    })
  }
}
</script>

<template>
  <div class="premium-addon-box">
    <div class="premium-addon-box__header">
      <span class="premium-addon-box__title">
        {{
          $t('pricing.addons.effective_balance', {
            amount: extraEffectiveBalance,
          })
        }}
      </span>
      <span class="premium-addon-box__title-detail">
        {{ $t('pricing.per_min_validator_deposit', {
          amount: formatFiatCurrency(pricePerValidator, {
            minimumFractionDigits: 5,
          }),
          old_validator_max_effective_balance: oldMaxEffectiveBalanceWithUnit,
        }) }}
      </span>
    </div>

    <hr>

    <div class="premium-addon-box__body">
      <div class="premium-addon-box__price">
        {{ formatFiatCurrency(isPaymentYearly ? totalYearlyPricePerMonth : totalMonthlyPrice) }}
      </div>
      <div class="premium-addon-box__price-description">
        {{ $t("pricing.per_month") }}
        {{ isPaymentYearly ? $t("pricing.amount_per_year", { amount: formatFiatCurrency(totalYearlyPrice) }) : '' }} *
      </div>
      <div
        v-if="isPaymentYearly"
        class="premium-addon-box__info-badge"
      >
        <span>
          {{ $t("pricing.savings", {
            amount: formatFiatCurrency(yearlySubscriptionSavings, { maximumFractionDigits: 0 }),
          }) }}
        </span>
        <BcTooltip
          position="top"
          :fit-content="true"
        >
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="premium-addon-box__info-tooltip">
              {{
                $t("pricing.savings_tooltip", {
                  monthly: formatFiatCurrency(totalMonthlyPrice),
                  monthly_yearly: formatFiatCurrency(totalYearlyPricePerMonth),
                })
              }}
            </div>
          </template>
        </BcTooltip>
      </div>
      <div v-if="addonSubscriptionCount">
        {{
          $t("pricing.addons.currently_active", {
            amount: addonSubscriptionCount,
          })
        }}
      </div>
      <div
        v-else
        class="premium-addon-box__subscription-form"
      >
        <fieldset
          aria-labelledby="subscription-count-row-label"
          class="premium-addon-box__subscription-form-count-row"
        >
          <BcScreenreaderOnly
            id="subscription-count-row-label"
            tag="legend"
          >
            {{ $t('pricing.addons.select_quantity') }}
          </BcScreenreaderOnly>
          <BcButton
            class="premium-addon-box__subscription-counter-button"
            :is-disabled="quantity <= 1"
            @click="quantity -= 1"
          >
            <FontAwesomeIcon :icon="faMinus" />
            <BcScreenreaderOnly>{{ $t('pricing.addons.button.decrease_quantity') }}</BcScreenreaderOnly>
          </BcButton>
          <BcInputNumber
            v-model="quantity"
            input-mode="numeric"
            :min="1"
            :max="maximumQuantity"
            :aria-label="$t('pricing.addons.selected_quantity')"
            input-width="2.75rem"
          />
          <BcButton
            class="premium-addon-box__subscription-counter-button"
            :is-disabled="isQuantityLimitReached"
            @click="quantity += 1"
          >
            <FontAwesomeIcon :icon="faPlus" />
            <BcScreenreaderOnly>{{ $t('pricing.addons.button.increase_quantity') }}</BcScreenreaderOnly>
          </BcButton>
        </fieldset>

        <span class="premium-addon-box__contact-support-text">
          <BcTranslation
            v-if="isQuantityLimitReached"
            keypath="pricing.addons.contact_support.template"
            linkpath="pricing.addons.contact_support._link"
            to="https://dsc.gg/beaconchain"
          />
        </span>

        <BcButton
          type="submit"
          :is-disabled="isDisabledSubmitButton"
          class="premium-addon-box__submit-button"
          @click="handleSubmitPurchase"
        >
          {{ isLoggedIn
            ? addonSubscriptionCount > 0
              ? $t('pricing.addons.button.manage_addon')
              : $t('pricing.addons.button.select_addon')
            : $t('pricing.get_started') }}
        </BcButton>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.premium-addon-box {
  width: calc(194px + 2.5rem);
  height: 100%;
  background-color: var(--container-background);
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  text-align: center;
  padding: 1.25rem;

  &__title {
    text-wrap: balance;
    display: inline-block;
    font-size: 1.25rem;
    white-space: pre-line;
    margin-bottom: 0.5rem;
  }

  &__title-detail {
    font-size: 1rem;
    color: var(--text-color-discreet);
  }

  hr {
    border-color: var(--container-border-color)
  }

  &__body {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 165px;
    margin: auto;
  }

  &__price {
    font-size: 1.75rem;
  }

  &__price-description {
    color: var(--text-color-discreet);
    font-size: 1rem;
    margin-bottom: 1rem;
  }

  &__info-badge {
    display: flex;
    width: 100%;
    background-color: var(--subcontainer-background);
    font-size: 0.7rem;
    padding: 0.25rem 0.15rem;
    border-radius: 4px;
    gap: 0.25rem;
    justify-content: center;
    margin-bottom: 0.75rem;
  }

  &__info-tooltip {
    width: 150px;
    text-align: left;
  }

  &__subscription-form {
    width: 100%;
  }

  &__subscription-form-count-row {
    display: flex;
    justify-content: space-between;
    border: none;
    padding: 0;
    margin: 0 -0.25rem;
  }

  &__subscription-counter-button {
    width: 36px;
    height: 36px;
    padding: 0.5rem;
  }

  &__contact-support-text {
    display: block;
    font-size: 0.75rem;
    height: 2.25rem;
  }

  &__submit-button {
    width: 100%;
    margin: 0
  }
}
</style>
