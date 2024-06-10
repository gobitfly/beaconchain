export interface StripeProvider{
  stripeInit: (stripePulicKey: string) => Promise<void>,
  stripeCustomerPortal: () => Promise<void>,
  stripePurchase: (priceId: number, amount: number) => Promise<void>,
  isStripeDisabled: Ref<boolean>
}
