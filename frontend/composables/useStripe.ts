import { inject, warn } from 'vue'
import type { StripeProvider } from '~/types/stripe'

export function useStripe() {
  const stripe = inject<StripeProvider>('stripe')

  const stripeInit = async (stripePulicKey: string) => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    await stripe.stripeInit(stripePulicKey)
  }

  const stripeCustomerPortal = async () => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    await stripe.stripeCustomerPortal()
  }

  const stripePurchase = async (priceId: string, amount: number) => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    await stripe.stripePurchase(priceId, amount)
  }

  const isStripeDisabled = computed(() => {
    if (!stripe) {
      warn('stripe provider not injected')
      return
    }

    return stripe?.isStripeDisabled.value
  })

  return { stripeInit, stripeCustomerPortal, stripePurchase, isStripeDisabled }
}
