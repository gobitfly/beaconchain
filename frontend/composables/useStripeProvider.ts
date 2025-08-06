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

  const { apiClient } = useRuntimeConfig().public
  const csrfToken = ref()
  const stripeInit = async (stripePulicKey: string) => {
    if (stripePulicKey === '') {
      return
    }
    await $fetch(`${apiClient}/pricing`, {
      onResponse({
        response,
      }) {
        csrfToken.value = response.headers.get('x-csrf-token') ?? ''
      },
    })
    stripe.value = await loadStripe(stripePulicKey)
  }

  const stripeCustomerPortal = async () => {
    if (isStripeDisabled.value) {
      return
    }

    isStripeProcessing.value = true

    const res = await fetch<StripeCustomerPortal>(
      'STRIPE_CUSTOMER_PORTAL',
      {
        baseURL: stripeBaseUrl,
        body: JSON.stringify({ returnURL: window.location.href }),
        headers: {
          'x-csrf-token': csrfToken.value,
        },
      },
    )

    await navigateTo(res?.url, { external: true })

    isStripeProcessing.value = false
  }

  const stripePurchase = async (priceId: string, amount: number) => {
    if (isStripeDisabled.value) {
      return
    }

    isStripeProcessing.value = true

    const res = await fetch<StripeCreateCheckoutSession>(
      'STRIPE_CHECKOUT_SESSION',
      {
        baseURL: stripeBaseUrl,
        body: JSON.stringify({
          addonQuantity: amount,
          priceId,
          promotionCode: promoCode,
        }),
        headers: {
          'x-csrf-token': csrfToken.value,
        },
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
