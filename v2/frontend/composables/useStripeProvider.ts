import { provide } from 'vue'
import type { StripeProvider } from '~/types/stripe'
import type { StripeCustomerPortal, StripeCreateCheckoutSession } from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'

export function useStripeProvider () {
  const { fetch } = useCustomFetch()

  const isStripeProcessing = ref(false)

  // TODO: Testcode, remove
  const sleep = (milliseconds: number): Promise<void> => {
    return new Promise<void>((resolve) => {
      setTimeout(resolve, milliseconds)
    })
  }

  const stripeCustomerPortal = async () => {
    isStripeProcessing.value = true

    await (sleep(1000)) // TODO: Test code, remove

    const res = await fetch<StripeCustomerPortal>(API_PATH.STRIPE_CUSTOMER_PORTAL, {
      body: JSON.stringify({ returnURL: window.location.href })
    })
    isStripeProcessing.value = false

    window.open(res?.url, '_blank')
  }

  const stripePurchase = async (priceId: number, amount: number) => {
    isStripeProcessing.value = true

    await (sleep(1000)) // TODO: Test code, remove

    const res = await fetch<StripeCreateCheckoutSession>(API_PATH.STRIPE_CHECKOUT_SESSION, {
      body: JSON.stringify({ priceIde: priceId, addonQuantity: amount })
    })
    isStripeProcessing.value = false

    console.log('StripeCreateCheckoutSession res', res)

    /*
    TODO:
    Use https://js.stripe.com/v3/ and then
    stripe.redirectToCheckout({ sessionId: d.sessionId }).then(handleResult).catch(err => {
      console.error("error redirecting to stripe checkout", err)
    */
  }

  provide<StripeProvider>('stripe', { stripeCustomerPortal, stripePurchase, isStripeProcessing })
}
