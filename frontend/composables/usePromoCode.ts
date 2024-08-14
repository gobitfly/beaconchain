export function usePromoCode() {
  const route = useRoute()

  const promoCode = route?.query?.promoCode

  return {
    promoCode,
  }
}
