export interface StripeProvider {
  isStripeDisabled: Ref<boolean>
  stripeCustomerPortal: () => Promise<void>
  stripeInit: (stripePulicKey: string) => Promise<void>
  stripePurchase: (priceId: string, amount: number) => Promise<void>
}
