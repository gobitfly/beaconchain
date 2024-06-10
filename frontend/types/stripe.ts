export interface StripeProvider{
  stripeInit: (stripePulicKey: string) => Promise<void>,
  stripeCustomerPortal: () => Promise<void>,
  stripePurchase: (priceId: string, amount: number) => Promise<void>,
  isStripeDisabled: Ref<boolean>
}
