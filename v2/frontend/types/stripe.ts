export interface StripeProvider{
  stripeCustomerPortal: () => Promise<void>,
  stripePurchase: (priceId: number, amount: number) => Promise<void>,
  isStripeProcessing: Ref<boolean>
}
