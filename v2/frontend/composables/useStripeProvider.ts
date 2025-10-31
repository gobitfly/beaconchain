import {
  provide, warn,
} from 'vue'
import {
  loadStripe, type Stripe,
} from '@stripe/stripe-js'
import type { StripeProvider } from '~/types/stripe'
import type {
  StripeCreateCheckoutSession,
  StripeCustomerPortal,
} from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'

export function useStripeProvider() {
  const { fetch } = useCustomFetch()
  const { promoCode } = usePromoCode()
  const { public: { stripeBaseUrl } } = useRuntimeConfig()

  const stripe = ref<null | Stripe>(null)

  const isStripeProcessing = ref(false)

  const isStripeDisabled = computed(() => {
    return (
      stripe === null || stripe.value === undefined || isStripeProcessing.value
    )
  })

  const stripeInit = async (stripePulicKey: string) => {
    if (stripePulicKey === '') {
      return
    }

    stripe.value = await loadStripe(stripePulicKey)
  }

  const stripeCustomerPortal = async () => {
    if (isStripeDisabled.value) {
      return
    }

    isStripeProcessing.value = true

    const res = await fetch<StripeCustomerPortal>(
      API_PATH.STRIPE_CUSTOMER_PORTAL,
      {
        baseURL: stripeBaseUrl,
        body: JSON.stringify({ returnURL: window.location.href }),
      },
    )

    window.open(res?.url, '_blank')

    isStripeProcessing.value = false
  }

  const stripePurchase = async (priceId: string, amount: number) => {
    if (isStripeDisabled.value) {
      return
    }

    isStripeProcessing.value = true

    const res = await fetch<StripeCreateCheckoutSession>(
      API_PATH.STRIPE_CHECKOUT_SESSION,
      {
        baseURL: stripeBaseUrl,
        body: JSON.stringify({
          addonQuantity: amount,
          priceId,
          promotionCode: promoCode,
        }),
      },
    )

    if (res.sessionId) {
      stripe.value!.redirectToCheckout({ sessionId: res.sessionId }) // stripe.value! checked via isStripeDisabled.value
    }
    else {
      warn('StripeCreateCheckoutSession error', res)
    }

    isStripeProcessing.value = false
  }

  provide<StripeProvider>('stripe', {
    isStripeDisabled,
    stripeCustomerPortal,
    stripeInit,
    stripePurchase,
  })

  return { stripeInit }
}
