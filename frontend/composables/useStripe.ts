import { inject, warn } from 'vue'
import type { StripeProvider } from '~/types/stripe'

export function useStripe () {
  const stripe = inject<StripeProvider>('stripe')

  const stripeCustomerPortal = async () => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    await stripe.stripeCustomerPortal()
  }

  const stripePurchase = async (priceId: number, amount: number) => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    await stripe.stripePurchase(priceId, amount)
  }

  const isStripeProcessing = computed(() => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    return stripe?.isStripeProcessing.value
  })

  return { stripeCustomerPortal, stripePurchase, isStripeProcessing }
}
